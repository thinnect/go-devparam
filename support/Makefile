# Makefile for building packages

# Determine version for the windows build
DEVP_VER=$(shell head -n 1 debian/changelog | sed -n "s/mist-device-parameters (\(.*\)).*/\1/p")

amd64:
	dpkg-buildpackage --host-arch amd64 -b --no-sign

armel:
	dpkg-buildpackage --host-arch armel -b --no-sign

armhf:
	dpkg-buildpackage --host-arch armhf -b --no-sign

win64:
	make -C ../cmd/deviceparameter win64
	make -C ../cmd/deviceparameters win64
	zip -j mist-device-parameters_$(DEVP_VER).zip ../cmd/deviceparameter/build/windows-amd64/deviceparameter.exe ../cmd/deviceparameters/build/windows-amd64/deviceparameters.exe
	mv mist-device-parameters_$(DEVP_VER).zip ../
