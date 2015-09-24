#!/usr/bin/env bash

if [ "$TRAVIS_OS_NAME" == "linux" ]; then

	echo "Package installs configured in .travis.yml"

	# Add the following to .travis.yml:

	# sudo: false
	#
	# addons:
	#   apt:
	#     sources:
	#       - ubuntu-sdk-team
	#     packages:
	#       - qtbase5-private-dev
	#       - qtdeclarative5-private-dev

elif [ "$TRAVIS_OS_NAME" == "osx" ]; then

	brew install qt5
	brew link --force qt5

else

	echo "BUILD NOT CONFIGURED: $TRAVIS_OS_NAME"
	exit 1

fi
