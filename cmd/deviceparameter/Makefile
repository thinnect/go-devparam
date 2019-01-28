# Makefile for embedding build info into the executable

BUILD_DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_DISTRO = $(shell lsb_release -sd)

all: deviceparameter manual

win32: export GOOS=windows
win32: export GOARCH=386
win32: deviceparameter.exe

win64: export GOOS=windows
win64: export GOARCH=amd64
win64: deviceparameter.exe

raspberry: export GOOS=linux
raspberry: export GOARCH=arm
raspberry: export GOARM=6
raspberry: deviceparameter

raspberry2: export GOOS=linux
raspberry2: export GOARCH=arm
raspberry2: export GOARM=7
raspberry2: deviceparameter

deviceparameter:
	go build -o deviceparameter -ldflags "-X 'main.ApplicationBuildDate=$(BUILD_DATE)' -X 'main.ApplicationBuildDistro=$(BUILD_DISTRO)'"

deviceparameter.exe:
	go build -o deviceparameter.exe -ldflags "-X 'main.ApplicationBuildDate=$(BUILD_DATE)' -X 'main.ApplicationBuildDistro=$(BUILD_DISTRO)'"

deviceparameter.1.gz:
	ronn --roff README.md
	mv README.1 deviceparameter.1
	gzip deviceparameter.1

manual: deviceparameter.1.gz

install: deviceparameter manual
	install -m 0755 deviceparameter /usr/local/bin
	install -m 0644 deviceparameter /usr/local/man/man1

clean:
	rm -f deviceparameter
	rm -f deviceparameter.1.gz

.PHONY: clean
