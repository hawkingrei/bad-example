load(
    "@io_bazel_rules_go//proto:def.bzl",
    "go_proto_library",
)

package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

proto_library(
    name = "model_proto",
    srcs = ["model.proto"],
    import_prefix = "go-common/app/service/openplatform/pgc-season/api/grpc/episode/v1",
    strip_import_prefix = "",
    tags = ["automanaged"],
    deps = ["@gogo_special_proto//github.com/gogo/protobuf/gogoproto"],
)

go_proto_library(
    name = "model_go_proto",
    compilers = ["@io_bazel_rules_go//proto:go_grpc"],
    importpath = "go-common/app/service/openplatform/pgc-season/api/grpc/episode/v1",
    proto = ":model_proto",
    tags = ["automanaged"],
    deps = ["@com_github_gogo_protobuf//gogoproto:go_default_library"],
)

proto_library(
    name = "service_proto",
    srcs = ["service.proto"],
    import_prefix = "go-common/app/service/openplatform/pgc-season/api/grpc/episode/v1",
    strip_import_prefix = "",
    tags = ["automanaged"],
    deps = [
        ":model_proto",
        "@gogo_special_proto//github.com/gogo/protobuf/gogoproto",
    ],
)

go_proto_library(
    name = "service_go_proto",
    compilers = ["@io_bazel_rules_go//proto:go_grpc"],
    importpath = "go-common/app/service/openplatform/pgc-season/api/grpc/episode/v1",
    proto = ":service_proto",
    tags = [
        "automanaged",
    ],
    deps = [
        ":model_go_proto",
        "@com_github_gogo_protobuf//gogoproto:go_default_library",
    ],
)

go_library(
    name = "go_default_library",
    srcs = [],
    embed = [
        ":model_go_proto",
        ":service_go_proto",
    ],
    importpath = "go-common/app/service/openplatform/pgc-season/api/grpc/episode/v1",
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_gogo_protobuf//proto:go_default_library",
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
