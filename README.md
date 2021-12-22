# go-devparam

Go library and applications for accessing device parameters using the
deviceparameters protocol: <https://github.com/thinnect/tos-devparam>.

`deviceparameter`
A basic utility for dealing with individual nodes and parameters.
See the [deviceparameter README](cmd/deviceparameter/README.md) for details.

`deviceparameters`
A utility for dealing with several parameters on multiple nodes.
See the [deviceparameters README](cmd/deviceparameters/README.md) for details.

# Building

Enter `cmd/deviceparameter` or `cmd/deviceparameters` and execute `make` to
see supported targets. Both applications can be cross-compiled for Windows and
for use on ARM based Linux platforms.

Packaged versions can be built from the support directory, see the
[support/Makefile](support/Makefile) for available options.

Building the deb packages requires `ronn`, which can be installed with ruby's
`gem` or can be obtained from <https://github.com/rtomayko/ronn>.
