"""Rule to build a RustProtoLibInfo and lib.rs for generated proto sources."""

load("//:defs.bzl", "ProtoCompileInfo")

def _strip_extension(f):
    return f.basename[:-len(f.extension) - 1]

def _rust_proto_lib_impl(ctx):
    """Generate a lib.rs file for the crates."""
    compilation = ctx.attr.compilation[ProtoCompileInfo]

    lib_rs = ctx.actions.declare_file(ctx.attr.name + "_lib.rs")

    b = compilation.output_dirs.to_list()[0]
    f = ctx.attr.lib.files.to_list()[0]

    if ctx.attr.lib != None:
        ctx.actions.run_shell(
            inputs = [f],
            outputs = [lib_rs],
            command = "sed 's/include!(\"/include!(\"%s\\//g' '%s' > '%s'" % (b.basename, f.path, lib_rs.path)
        )
    else:
        # TODO: restore old include behavior
        ctx.actions.write(lib_rs, "")


    return [DefaultInfo(
        files = depset([lib_rs]),
    )]

rust_proto_lib = rule(
    implementation = _rust_proto_lib_impl,
    attrs = {
        "compilation": attr.label(
            providers = [ProtoCompileInfo],
            mandatory = True,
        ),
        "lib": attr.label(
            allow_single_file=True,
            mandatory = False,
        ),
    },
)
