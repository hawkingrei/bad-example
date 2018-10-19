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
        "//app/admin/main/vip/model:all-srcs",
        "//app/interface:all-srcs",
        "//app/service:all-srcs",
        "//app/service/main/coupon/model:all-srcs",
    ],
    tags = ["automanaged"],
)
