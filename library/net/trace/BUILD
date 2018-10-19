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
        "//library/net/trace/proto:all-srcs",
    ],
    tags = ["automanaged"],
)

go_library(
    name = "go_default_library",
    srcs = [
        "propagation.go",
        "report.go",
        "sample.go",
        "tag.go",
        "trace.go",
        "trace_v2.go",
        "tracer.go",
        "tracer_v2.go",
        "util.go",
    ],
    importpath = "go-common/library/net/trace",
    tags = ["automanaged"],
    deps = [
        "//library/conf/dsn:go_default_library",
        "//library/conf/env:go_default_library",
        "//library/net/metadata:go_default_library",
        "//library/net/trace/proto:go_default_library",
        "//library/stat/counter:go_default_library",
        "//library/time:go_default_library",
        "//vendor/github.com/dgryski/go-farm:go_default_library",
        "//vendor/github.com/pkg/errors:go_default_library",
        "//vendor/github.com/satori/go.uuid:go_default_library",
        "@org_golang_google_grpc//metadata:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "report_test.go",
        "sample_test.go",
        "trace_test.go",
        "tracer_test.go",
    ],
    embed = [":go_default_library"],
    rundir = ".",
    tags = ["automanaged"],
    deps = [
        "//library/net/metadata:go_default_library",
        "//library/time:go_default_library",
        "//vendor/github.com/pkg/errors:go_default_library",
        "//vendor/github.com/stretchr/testify/assert:go_default_library",
        "@org_golang_google_grpc//metadata:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    tags = ["automanaged"],
    deps = ["//library/net/trace:go_default_library"],
)
