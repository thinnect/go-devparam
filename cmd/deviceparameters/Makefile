# Makefile for embedding build info into the executable

BUILD_DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_DISTRO = $(shell lsb_release -sd)

all: deviceparameters

win32: export GOOS=windows
win32: export GOARCH=386
win32: deviceparameters.exe

win64: export GOOS=windows
win64: export GOARCH=amd64
win64: deviceparameters.exe

raspberry: export GOOS=linux
raspberry: export GOARCH=arm
raspberry: export GOARM=6
raspberry: deviceparameters

raspberry2: export GOOS=linux
raspberry2: export GOARCH=arm
raspberry2: export GOARM=7
raspberry2: deviceparameters

deviceparameters:
	go build -o deviceparameters -ldflags "-X 'main.ApplicationBuildDate=$(BUILD_DATE)' -X 'main.ApplicationBuildDistro=$(BUILD_DISTRO)'"

deviceparameters.exe:
	go build -o deviceparameters.exe -ldflags "-X 'main.ApplicationBuildDate=$(BUILD_DATE)' -X 'main.ApplicationBuildDistro=$(BUILD_DISTRO)'"

deviceparameters.1.gz:
	ronn --roff README.md
	mv README.1 deviceparameters.1
	gzip deviceparameters.1

manual: deviceparameters.1.gz

install: deviceparameters manual
	install -m 0755 deviceparameters /usr/local/bin
	install -m 0644 deviceparameters /usr/local/man/man1
