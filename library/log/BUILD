load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

package(default_visibility = ["//visibility:public"])

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//library/log/anticheat:all-srcs",
        "//library/log/infoc:all-srcs",
    ],
    tags = ["automanaged"],
)

go_library(
    name = "go_default_library",
    srcs = [
        "agent.go",
        "doc.go",
        "dsn.go",
        "file.go",
        "handler.go",
        "level.go",
        "log.go",
        "logrus.go",
        "pattern.go",
        "stdout.go",
        "util.go",
        "verbose.go",
    ],
    importpath = "go-common/library/log",
    tags = ["automanaged"],
    deps = [
        "//library/conf/dsn:go_default_library",
        "//library/conf/env:go_default_library",
        "//library/net/metadata:go_default_library",
        "//library/net/trace:go_default_library",
        "//library/stat/prom:go_default_library",
        "//library/time:go_default_library",
        "//vendor/github.com/alecthomas/log4go:go_default_library",
        "//vendor/github.com/pkg/errors:go_default_library",
        "//vendor/github.com/sirupsen/logrus:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "dsn_test.go",
        "log_test.go",
        "pattern_test.go",
    ],
    embed = [":go_default_library"],
    rundir = ".",
    tags = ["automanaged"],
    deps = [
        "//library/net/metadata:go_default_library",
        "//library/time:go_default_library",
        "//vendor/github.com/stretchr/testify/assert:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    tags = ["automanaged"],
    deps = ["//library/log:go_default_library"],
)
