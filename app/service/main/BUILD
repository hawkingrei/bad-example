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
        "//app/service/main/account:all-srcs",
        "//app/service/main/member:all-srcs",
        "//app/service/main/relation:all-srcs",
        "//app/service/main/usersuit:all-srcs",
        "//app/service/main/vip:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
