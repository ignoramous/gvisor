// Copyright 2018 The gVisor Authors.
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

// Package main implements a tool that can save and replay messages from
// issued from remote.Remote.
package main

import (
	"context"
	"os"

	"github.com/google/subcommands"
	"gvisor.dev/gvisor/runsc/flag"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(&save{}, "")
	subcommands.Register(&replay{}, "")
	flag.CommandLine.Parse(os.Args[1:])
	os.Exit(int(subcommands.Execute(context.Background())))
}
