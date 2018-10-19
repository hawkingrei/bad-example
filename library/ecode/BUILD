package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_test",
    "go_library",
)

go_test(
    name = "go_default_test",
    srcs = ["ecode_test.go"],
    embed = [":go_default_library"],
    rundir = ".",
    tags = ["automanaged"],
)

go_library(
    name = "go_default_library",
    srcs = [
        "bbq.go",
        "ecode.go",
        "ep.go",
        "live.go",
        "main.go",
        "open.go",
    ],
    importpath = "go-common/library/ecode",
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/pkg/errors:go_default_library"],
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    tags = ["automanaged"],
    deps = [
        "//library/ecode:go_default_library",
        "//vendor/github.com/pkg/errors:go_default_library",
    ],
)

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
        "//library/ecode/pb:all-srcs",
        "//library/ecode/tip:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
