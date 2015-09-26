#!/usr/bin/env bash

if [ "$TRAVIS_OS_NAME" == "linux" ]; then

    sudo apt-get -qy install qtbase5-private-dev libqt5opengl5 libqt5opengl5-dev

elif [ "$TRAVIS_OS_NAME" == "osx" ]; then

	brew install qt5
	brew link --force qt5 --with-developer

else

	echo "BUILD NOT CONFIGURED: $TRAVIS_OS_NAME"
	exit 1

fi
