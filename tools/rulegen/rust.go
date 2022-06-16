package main

var rustWorkspaceTemplate = mustTemplate(`load("@rules_proto_grpc//{{ .Lang.Dir }}:repositories.bzl", rules_proto_grpc_{{ .Lang.Name }}_repos = "{{ .Lang.Name }}_repos")

rules_proto_grpc_{{ .Lang.Name }}_repos()

load("@rules_rust//rust:repositories.bzl", "rust_repositories")

rust_repositories()`)

var rustLibraryRuleTemplateString = `load("//{{ .Lang.Dir }}:{{ .Rule.Base }}_{{ .Rule.Kind }}_compile.bzl", "{{ .Rule.Base }}_{{ .Rule.Kind }}_compile")
load("//internal:compile.bzl", "proto_compile_attrs")
load("//{{ .Lang.Dir }}:rust_proto_lib.bzl", "rust_proto_lib")
load("@rules_rust//rust:defs.bzl", "rust_library")

def {{ .Rule.Name }}(name, **kwargs):  # buildifier: disable=function-docstring
    # Compile protos
    name_pb = name + "_pb"
    name_lib = name + "_lib"
    {{ .Rule.Base }}_{{ .Rule.Kind }}_compile(
        name = name_pb,
        {{ .Common.ArgsForwardingSnippet }}
    )
`

var rustProstProtoLibraryRuleTemplate = mustTemplate(rustLibraryRuleTemplateString + `
    # Create lib file
    rust_proto_lib(
        name = name_lib,
        compilation = name_pb,
    )

    # Create {{ .Rule.Base }} library
    rust_library(
        name = name,
        edition = "2018",
        srcs = [name_pb, name_lib],
        deps = kwargs.get("prost_deps", [Label("@crate_index//:prost"), Label("@crate_index//:prost-types")]) + kwargs.get("deps", []),
        proc_macro_deps = [kwargs.get("prost_derive_dep", Label("@crate_index//:prost-derive"))],
        visibility = kwargs.get("visibility"),
        tags = kwargs.get("tags"),
    )
`)

var rustTonicGrpcLibraryRuleTemplate = mustTemplate(rustLibraryRuleTemplateString + `
    # Create lib file
    rust_proto_lib(
        name = name_lib,
        compilation = name_pb,
    )

    # Create {{ .Rule.Base }} library
    rust_library(
        name = name,
        edition = "2018",
        srcs = [name_pb, name_lib],
        deps = kwargs.get("prost_deps", [Label("@crate_index//:prost"), Label("@crate_index//:prost-types")]) +
          [kwargs.get("tonic_dep", Label("@crate_index//:tonic"))] +
          kwargs.get("deps", []),
        proc_macro_deps = [kwargs.get("prost_derive_dep", Label("@crate_index//:prost-derive"))],
        visibility = kwargs.get("visibility"),
        tags = kwargs.get("tags"),
    )
`)

// For rust, produce one library for all protos, since they are all in the same crate
var rustProtoLibraryExampleTemplate = mustTemplate(`load("@rules_proto_grpc//{{ .Lang.Dir }}:defs.bzl", "{{ .Rule.Name }}")

{{ .Rule.Name }}(
    name = "proto_{{ .Rule.Base }}_{{ .Rule.Kind }}",
    protos = [
        "@rules_proto_grpc//example/proto:person_proto",
        "@rules_proto_grpc//example/proto:place_proto",
        "@rules_proto_grpc//example/proto:thing_proto",
    ],
)`)

var rustGrpcLibraryExampleTemplate = mustTemplate(`load("@rules_proto_grpc//{{ .Lang.Dir }}:defs.bzl", "{{ .Rule.Name }}")

{{ .Rule.Name }}(
    name = "greeter_{{ .Rule.Base }}_{{ .Rule.Kind }}",
    protos = [
        "@rules_proto_grpc//example/proto:greeter_grpc",
        "@rules_proto_grpc//example/proto:thing_proto",
    ],
)`)

var rustProstLibraryRuleAttrs = append(append([]*Attr{}, libraryRuleAttrs...), []*Attr{
	&Attr{
		Name:      "prost_deps",
		Type:      "label_list",
		Default:   `["@crate_index//:prost", "@crate_index//:prost-types"]`,
		Doc:       "The prost dependencies that the rust library should depend on.",
		Mandatory: false,
	},
	&Attr{
		Name:      "prost_derive_dep",
		Type:      "label",
		Default:   `@crate_index//:prost-derive`,
		Doc:       "The prost-derive dependency that the rust library should depend on.",
		Mandatory: false,
	},
}...)

var rustTonicLibraryRuleAttrs = append(append([]*Attr{}, rustProstLibraryRuleAttrs...), []*Attr{
	&Attr{
		Name:      "tonic_dep",
		Type:      "label",
		Default:   `@crate_index//:tonic`,
		Doc:       "The tonic dependency that the rust library should depend on.",
		Mandatory: false,
	},
}...)

func makeRust() *Language {
	return &Language{
		Dir:               "rust",
		Name:              "rust",
		DisplayName:       "Rust",
		Notes:             mustTemplate("Rules for generating Rust protobuf and gRPC ``.rs`` files and libraries using `prost <https://github.com/tokio-rs/prost>`_ and `tonic <https://github.com/hyperium/tonic>`_. Libraries are created with ``rust_library`` from `rules_rust <https://github.com/bazelbuild/rules_rust>`_. Requires ``--experimental_proto_descriptor_sets_include_source_info`` to be set for the build."),
		Flags:             commonLangFlags,
		SkipTestPlatforms: []string{"windows", "macos"}, // CI has no rust toolchain for windows and is broken on mac
		Rules: []*Rule{
			&Rule{
				Name:             "rust_prost_proto_compile",
				Base:             "rust_prost",
				Kind:             "proto",
				Implementation:   compileRuleTemplate,
				Plugins:          []string{"//rust:rust_prost_plugin"},
				WorkspaceExample: rustWorkspaceTemplate,
				BuildExample:     protoCompileExampleTemplate,
				Doc:              "Generates Rust protobuf ``.rs`` files using prost",
				Attrs:            compileRuleAttrs,
			},
			&Rule{
				Name:             "rust_tonic_grpc_compile",
				Base:             "rust_tonic",
				Kind:             "grpc",
				Implementation:   compileRuleTemplate,
				Plugins:          []string{"//rust:rust_prost_plugin", "//rust:rust_tonic_plugin"},
				WorkspaceExample: rustWorkspaceTemplate,
				BuildExample:     grpcCompileExampleTemplate,
				Doc:              "Generates Rust protobuf and gRPC ``.rs`` files using prost and tonic",
				Attrs:            compileRuleAttrs,
			},
			&Rule{
				Name:             "rust_prost_proto_library",
				Base:             "rust_prost",
				Kind:             "proto",
				Implementation:   rustProstProtoLibraryRuleTemplate,
				WorkspaceExample: rustWorkspaceTemplate,
				BuildExample:     rustProtoLibraryExampleTemplate,
				Doc:              "Generates a Rust prost protobuf library using ``rust_library`` from ``rules_rust``",
				Attrs:            rustProstLibraryRuleAttrs,
			},
			&Rule{
				Name:             "rust_tonic_grpc_library",
				Base:             "rust_tonic",
				Kind:             "grpc",
				Implementation:   rustTonicGrpcLibraryRuleTemplate,
				WorkspaceExample: rustWorkspaceTemplate,
				BuildExample:     rustGrpcLibraryExampleTemplate,
				Doc:              "Generates a Rust prost protobuf and tonic gRPC library using ``rust_library`` from ``rules_rust``",
				Attrs:            rustTonicLibraryRuleAttrs,
			},
		},
	}
}
