// Copyright 2022 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/google/subcommands"
	"gvisor.dev/gvisor/pkg/atomicbitops"
	"gvisor.dev/gvisor/pkg/sentry/seccheck/checkers/remote/server"
	"gvisor.dev/gvisor/pkg/sentry/seccheck/checkers/remote/wire"
	"gvisor.dev/gvisor/runsc/flag"
)

// save implements subcommands.Command for the "save" command.
type save struct {
	endpoint string
	out      string
	prefix   string
}

// Name implements subcommands.Command.
func (*save) Name() string {
	return "save"
}

// Synopsis implements subcommands.Command.
func (*save) Synopsis() string {
	return "save trace sessions to files"
}

// Usage implements subcommands.Command.
func (*save) Usage() string {
	return `save [flags] - save trace sessions to files
`
}

// SetFlags implements subcommands.Command.
func (s *save) SetFlags(f *flag.FlagSet) {
	f.StringVar(&s.endpoint, "endpoint", "", "path to trace server endpoint to connect")
	f.StringVar(&s.out, "out", "./replay", "path to a directory where trace files will be saved")
	f.StringVar(&s.prefix, "prefix", "client-", "name to be prefixed to each trace file")
}

// Execute implements subcommands.Command.
func (s *save) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if f.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", f.Args())
		return subcommands.ExitUsageError
	}
	if len(s.endpoint) == 0 {
		fmt.Fprintf(os.Stderr, "--endpoint is required\n")
		return subcommands.ExitUsageError
	}
	_ = os.Remove(s.endpoint)

	server := newServer(s.endpoint, s.out, s.prefix)
	defer server.Close()

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "starting server: %v\n", err)
		return subcommands.ExitFailure
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	done := make(chan struct{})
	go func() {
		<-c
		fmt.Printf("Ctrl-C pressed, stopping.\n")
		done <- struct{}{}
	}()

	fmt.Printf("Listening on %q. Press ctrl-C to stop...\n", s.endpoint)
	<-done
	return subcommands.ExitSuccess
}

type saveServer struct {
	server.CommonServer
	dir         string
	prefix      string
	clientCount atomicbitops.Uint64
}

var _ server.ClientHandler = (*saveServer)(nil)

func newServer(endpoint, dir, prefix string) *saveServer {
	s := &saveServer{dir: dir, prefix: prefix}
	s.CommonServer.Init(endpoint, s)
	return s
}

// Start starts the server.
func (s *saveServer) Start() error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return err
	}
	return s.CommonServer.Start()
}

// NewClient creates a new file for the client and writes messages to it.
//
// The file format starts with a string signature to make it easy to check that
// it's a trace file. The signature is followed by a JSON configuration that
// contains information required to process the file. Next, there are a sequence
// of messages. Both JSON and messages are prefixed by an uint64 with their
// size.
//
// Ex:
//   signature <size>Config JSON [<size>message]*
func (s *saveServer) NewClient() (server.MessageHandler, error) {
	seq := s.clientCount.Add(1)
	filename := filepath.Join(s.dir, fmt.Sprintf("%s%04d", s.prefix, seq))
	fmt.Printf("New client connected, writing to: %q\n", filename)

	out, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	if _, err := out.Write([]byte(signature)); err != nil {
		return nil, err
	}

	handler := &msgHandler{out: out}

	cfg, err := json.Marshal(Config{Version: handler.Version()})
	if err != nil {
		return nil, err
	}
	if err := writeWithSize(out, cfg); err != nil {
		return nil, err
	}

	return handler, nil
}

type msgHandler struct {
	out          *os.File
	messageCount atomicbitops.Uint64
}

var _ server.MessageHandler = (*msgHandler)(nil)

// Version implements server.MessageHandler.
func (m *msgHandler) Version() uint32 {
	return wire.CurrentVersion
}

// Message saves the message to the client file.
func (m *msgHandler) Message(raw []byte, _ wire.Header, _ []byte) error {
	m.messageCount.Add(1)
	return writeWithSize(m.out, raw)
}

// Close closes the client file.
func (m *msgHandler) Close() {
	fmt.Printf("Closing client, wrote %d messages to %q\n", m.messageCount.Load(), m.out.Name())
	_ = m.out.Close()
}
