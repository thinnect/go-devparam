#!/bin/bash

RELEASE_DATE=$(date -u +'%Y%m%d')

OPTIONS+=" --type=debian"
OPTIONS+=" --default"     # answer yes to any questions
OPTIONS+=" --fstrans"     # simulate filesystem so root permissions not needed
OPTIONS+=" --install=no"  # just make, don't install the package
OPTIONS+=" --nodoc"       # as long as actual docs don't exist

OPTIONS+=" --pkgname=mist-device-parameters"
OPTIONS+=" --pkgversion=0.1.0"
OPTIONS+=" --pkgrelease=$RELEASE_DATE"
OPTIONS+=" --provides=mist-device-parameters"
OPTIONS+=" --pkglicense=MIT"
OPTIONS+=" --maintainer=somebody@thinnect.com"

OPTIONS+=" --exclude /usr/lib/python3"  # running lsb_release in a fresh container will make a pyc file there otherwise

checkinstall $OPTIONS
