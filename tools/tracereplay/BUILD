load("//tools:defs.bzl", "go_binary", "go_test")

package(licenses = ["notice"])

go_binary(
    name = "tracereplay",
    srcs = [
        "common.go",
        "main.go",
        "replay.go",
        "save.go",
    ],
    deps = [
        "//pkg/atomicbitops",
        "//pkg/log",
        "//pkg/sentry/seccheck/checkers/remote/server",
        "//pkg/sentry/seccheck/checkers/remote/wire",
        "//pkg/sentry/seccheck/points:points_go_proto",
        "//runsc/flag",
        "@com_github_google_subcommands//:go_default_library",
        "@org_golang_google_protobuf//proto:go_default_library",
        "@org_golang_x_sys//unix:go_default_library",
    ],
)

go_test(
    name = "tracereplay_test",
    srcs = ["main_test.go"],
    data = [
        "testdata/client-0001",
    ],
    library = ":tracereplay",
    deps = [
        "//pkg/test/testutil",
    ],
)
