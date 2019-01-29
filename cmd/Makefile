# Makefile for building and installing deviceparameter and deviceparameters
MAKEFILE_LOCATION := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

all: deviceparameter deviceparameters

deviceparameter:
	cd deviceparameter && make clean && make

deviceparameters:
	cd deviceparameters && make clean && make

install: deviceparameter deviceparameters
	install -d -m0755 /usr/share/man/man1
	install -m0755  $(MAKEFILE_LOCATION)/deviceparameter/deviceparameter /usr/bin
	install -m0664  $(MAKEFILE_LOCATION)/deviceparameter/deviceparameter.1.gz /usr/share/man/man1
	install -m0755  $(MAKEFILE_LOCATION)/deviceparameters/deviceparameters /usr/bin
	install -m0664  $(MAKEFILE_LOCATION)/deviceparameters/deviceparameters.1.gz /usr/share/man/man1

.PHONY: deviceparameter
.PHONY: deviceparameters
