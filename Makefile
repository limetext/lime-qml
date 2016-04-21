test:
	@go test -race ./main/...
build:
	cd main; go build
fmt:
	@go fmt ./main/...
license:
	@go run gen_license.go main

check_fmt:
ifneq ($(shell gofmt -l main),)
	$(error code not fmted, run make fmt. $(shell gofmt -l main))
endif

check_license:
ifneq ($(shell go run gen_license.go main),)
	$(error license is not added to all files, run make license)
endif

glide:
	go get -v -u github.com/Masterminds/glide
	glide install

travis:
ifeq ($(TRAVIS_OS_NAME),osx)
	brew update
	brew install oniguruma python3 qt5
	brew link --force qt5
endif

travis_test: export PKG_CONFIG_PATH += $(PWD)/vendor/github.com/limetext/rubex:$(GOPATH)/src/github.com/limetext/rubex
travis_test: test build
