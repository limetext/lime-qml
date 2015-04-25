#!/usr/bin/env bash

if [ "$TRAVIS_OS_NAME" == "linux" ]; then

	# install qml frontend dependencies
	sudo add-apt-repository -y ppa:ubuntu-sdk-team/ppa
	sudo apt-get update -qq
	sudo apt-get install -qq qtbase5-private-dev qtdeclarative5-private-dev

elif [ "$TRAVIS_OS_NAME" == "osx" ]; then

	brew install qt5
	brew link --force qt5

else

	echo "BUILD NOT CONFIGURED: $TRAVIS_OS_NAME"
	exit 1

fi
