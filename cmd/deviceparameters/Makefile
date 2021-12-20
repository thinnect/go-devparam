# Makefile for embedding build info into the executable

BUILD_DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_DISTRO = $(shell lsb_release -sd)

USE_UPX ?= 0
ifneq ($(USE_UPX),0)
	BUILD_PARTS := build compress-brute
else
	BUILD_PARTS := build
endif

# In this setup arm5=armel and arm6=armhf for widest compatibility
GOALS := amd64 arm5 armel arm6 armhf arm7 arm64 win64 clean
ifeq (,$(filter $(GOALS),$(MAKECMDGOALS)))
  $(error Build with make amd64/arm5/armel/arm6/armhf/arm7/arm64/win64)
endif

amd64:
amd64: export GOOS=linux
amd64: export GOARCH=amd64
amd64: export FLAVOUR=$(GOOS)-$(GOARCH)
amd64: $(BUILD_PARTS) manual

arm5: export GOOS=linux
arm5: export GOARCH=arm
arm5: export GOARM=5
arm5: export FLAVOUR=$(GOOS)-$(GOARCH)$(GOARM)
arm5: $(BUILD_PARTS) manual

armel: export GOOS=linux
armel: export GOARCH=arm
armel: export GOARM=5
armel: export FLAVOUR=$(GOOS)-armel
armel: $(BUILD_PARTS) manual

arm6: export GOOS=linux
arm6: export GOARCH=arm
arm6: export GOARM=6
arm6: export FLAVOUR=$(GOOS)-$(GOARCH)$(GOARM)
arm6: $(BUILD_PARTS) manual

armhf: export GOOS=linux
armhf: export GOARCH=arm
armhf: export GOARM=6
armhf: export FLAVOUR=$(GOOS)-armhf
armhf: $(BUILD_PARTS) manual

arm7: export GOOS=linux
arm7: export GOARCH=arm
arm7: export GOARM=7
arm7: export FLAVOUR=$(GOOS)-$(GOARCH)$(GOARM)
arm7: $(BUILD_PARTS) manual

arm64: export GOOS=linux
arm64: export GOARCH=arm64
arm64: export FLAVOUR=$(GOOS)-$(GOARCH)
arm64: $(BUILD_PARTS) manual

win64: export GOOS=windows
win64: export GOARCH=amd64
win64: export FLAVOUR=$(GOOS)-$(GOARCH)
win64: deviceparameters.exe

builddir: $(FLAVOUR)
	mkdir -p build/$(FLAVOUR)

# -s disable symbol table
# -w disable DWARF generation
build: builddir
	go build -o build/$(FLAVOUR)/deviceparameters -ldflags "-w -s -X 'main.ApplicationBuildDate=$(BUILD_DATE)' -X 'main.ApplicationBuildDistro=$(BUILD_DISTRO)'"

deviceparameters.exe:
	go build -o build/$(FLAVOUR)/deviceparameters.exe -ldflags "-w -s -X 'main.ApplicationBuildDate=$(BUILD_DATE)' -X 'main.ApplicationBuildDistro=$(BUILD_DISTRO)'"

# upx will make the binary much smaller
compress: build
	upx build/$(FLAVOUR)/deviceparameters

# but will take quite a while with --brute
compress-brute: build
	upx --brute build/$(FLAVOUR)/deviceparameters

build/$(FLAVOUR)/deviceparameters.1.gz:
	ronn --roff README.md
	mv README.1 deviceparameters.1
	gzip deviceparameters.1
	mv deviceparameters.1.gz build/$(FLAVOUR)/

manual: build/$(FLAVOUR)/deviceparameters.1.gz

clean:
	rm -Rf build

.PHONY: clean
