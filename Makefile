test:
	@go test -race ./main/...
build:
	cd main; go build
fmt:
	@go fmt ./main/...
license:
	@go run $(GOPATH)/src/github.com/limetext/tasks/gen_license.go -scan=main

check_fmt:
ifneq ($(shell gofmt -l main),)
	$(error code not fmted, run make fmt. $(shell gofmt -l main))
endif

check_license:
	@go run $(GOPATH)/src/github.com/limetext/tasks/gen_license.go -scan=main -check

tasks:
	go get -d -u github.com/limetext/tasks
glide:
	go get -v -u github.com/Masterminds/glide
	glide install

travis: tasks
ifeq ($(TRAVIS_OS_NAME),osx)
	brew update
	brew install oniguruma python3 qt5
	brew link --force qt5
endif

travis_test: export PKG_CONFIG_PATH += $(PWD)/vendor/github.com/limetext/rubex:$(GOPATH)/src/github.com/limetext/rubex
travis_test: test build
