package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_library(
    name = "go_default_library",
    srcs = [
        "app_info.go",
        "card.go",
        "coupon.go",
        "enum.go",
        "http.go",
        "invite.go",
        "model.go",
        "notice.go",
        "pendant.go",
        "realname.go",
        "regular.go",
        "relation.go",
        "reply.go",
        "usersuit.go",
        "vip.go",
        "web_info.go",
    ],
    importpath = "go-common/app/interface/main/account/model",
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
    deps = [
        "//app/service/main/account/model:go_default_library",
        "//app/service/main/member/model:go_default_library",
        "//app/service/main/relation/model:go_default_library",
        "//app/service/main/usersuit/model:go_default_library",
        "//app/service/main/vip/model:go_default_library",
        "//library/net/http/blademaster:go_default_library",
        "//library/time:go_default_library",
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
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["regular_test.go"],
    embed = [":go_default_library"],
    rundir = ".",
    tags = ["automanaged"],
    deps = ["//vendor/github.com/smartystreets/goconvey/convey:go_default_library"],
)
