# Published catalog candidate rehearsal

This isolated branch tests the unmerged profiler fix from
https://github.com/runtimeconditions/go-rc-profiler/pull/2 at exact commit
`2d48a3d2944b5dcf62df5284e93d05edc2c65e07` through the complete published-dependency
demo, including all five real KinD/Kratix acceptance cases. It is not a release.

The candidate library uses an explicit remote Go module replacement to
`github.com/ksteffe/go-rc-profiler@v0.0.0-20261021151955-2d48a3d2944b`.
The unmerged commit cannot yet be resolved under the upstream module path.
There is no local filesystem replacement.
Generation continues using the official checksum-pinned v0.6.0 executable, whose
output already matches the released bindings. Definitions come from genuine HTTPS
publication endpoints. No sibling repositories or fixture overrides are used in CI.

The previous bug reproducer now requires successful resolution. The producer's
no-signature-on-failure test supplies a genuinely invalid engine instead of relying
on the fixed identity bug. All original acceptance and integration checks remain.

The main migration branch and its released v0.6.0 dependency pin are unchanged.
A released profiler containing the fix is still needed for the final migration.
