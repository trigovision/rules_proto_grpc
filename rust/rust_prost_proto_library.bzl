"""Generated definition of rust_prost_proto_library."""

load("//rust:rust_prost_proto_compile.bzl", "rust_prost_proto_compile")
load("//internal:compile.bzl", "proto_compile_attrs")
load("//rust:rust_proto_lib.bzl", "rust_proto_lib")
load("@rules_rust//rust:defs.bzl", "rust_library")

def rust_prost_proto_library(name, **kwargs):  # buildifier: disable=function-docstring
    # Compile protos
    name_pb = name + "_pb"
    name_lib = name + "_lib"
    rust_prost_proto_compile(
        name = name_pb,
        **{
            k: v
            for (k, v) in kwargs.items()
            if k in proto_compile_attrs.keys()
        }  # Forward args
    )

    # Create lib file
    rust_proto_lib(
        name = name_lib,
        compilation = name_pb,
    )

    # Create rust_prost library
    rust_library(
        name = name,
        edition = "2018",
        srcs = [name_pb, name_lib],
        deps = kwargs.get("prost_deps", [Label("@crate_index//:prost"), Label("@crate_index//:prost-types")]) + kwargs.get("deps", []),
        proc_macro_deps = [kwargs.get("prost_derive_dep", Label("@crate_index//:prost-derive"))],
        visibility = kwargs.get("visibility"),
        tags = kwargs.get("tags"),
    )

