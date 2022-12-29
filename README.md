# bespoke
Run kustomize after substituting environment variables

Most folks run `kustomize` and then use `envsubst` to substitute 
environment variables. They do it in this order because
`kustomize` generates a single output stream that's easy to
feed into `envsubst`; the input to `kustomize` may be many files
spread over many directories. Unfortunately, that limits where
in the `kustomize` input variables may be used -- they can't be
used in resource names, patch target paths, etc.

(The `kustomize` maintainers don't like variables; the tool is
"opinionated" in this way and so the rest of us must find
workarounds.)

`bespoke` finds all the input files that `kustomize` would use,
copies them to a temporary directory while substituting variables
along the way, and then runs `kustomize` on that temporary
directory to generate the final output.

As much as possible, `bespoke` uses the `kustomize` API, but it
must make its own determination of what files or directories
are referenced from the `kustomization.yaml` file. As a result,
it may be a little fragile as `kustomize` evolves. For example,
many existing files use `bases` which is deprecated; when it's
finally removed from `kustomize`, a similar change must be made
in `bespoke`.

At the moment, `bespoke` doesn't really have any options; when
you run `bespoke build [dir]` it's the same as running 
`kustomize build` with `--enable-alpha-plugins`. The directory
is optional; the tool defaults to the current working directory
to find an input file.

`bespoke` uses a variant of a Go-based `envsubst` command that
allows for defaults in variables. It's a variant because the
default behavior of the original is to substitute "" when
variables are not set (or fail, with that option selected);
the variant will pass variables through as-is when not set,
so that they may be used, e.g., in scripts passed into 
Kubernetes pods (where, presumably, the variable is defined).

That is, if the input contains `$LC_PREFIX` or `${LC_PREFIX}`
and `LC_PREFIX` isn't set in the current environment, then the
output will have `${LC_PREFIX}` in it, rather than the empty
string.

