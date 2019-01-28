# go-devparam

Go library and applications for accessing device parameters using the
deviceparameters protocol: <https://github.com/thinnect/tos-devparam>.

`deviceparameter`
A basic utility for dealing with individual nodes and parameters.
See the [deviceparameter README](cmd/deviceparameter/README.md) for details.

`deviceparameters`
A utility for dealing with several parameters on multiple nodes.
See the [deviceparameters README](cmd/deviceparameters/README.md) for details.

# Dependencies
Go dependencies have been vendored as submodules under the vendor directory.

Building the _deb_ package requires `checkinstall` and `ronn`.

`ronn` can be obtained from <https://github.com/rtomayko/ronn>.

# Building

Check out all submodules, install `ronn` and `checkinstall`.

Enter `cmd/deviceparameter` or `cmd/deviceparameters` and execute `make` or
enter `cmd` and execute the `build-deb.sh` script.

Both applications can be cross-compiled for Windows and for use on ARM based
Linux platforms, though packaging only works on the native architecture. Take a
look at the Makefiles for details.
