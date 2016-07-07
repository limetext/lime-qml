all: submodule

precommit: fmt license test

build:
	cd main && go build

test:
	@go test -race ./main/...

run:
	cd main && ./main

clean:
	rm main/main main/debug.log

test_run: build
	cd main && ./main & export TASK_PID=$$! && sleep 10 && kill $$TASK_PID

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

submodule:
	git submodule update --init --recursive

glide:
	go get -v -u github.com/Masterminds/glide
	glide install

travis:
ifeq ($(TRAVIS_OS_NAME),osx)
	brew update
	brew install oniguruma python3 qt5
	brew link --force qt5
endif
travis: glide tasks

travis_test: export PKG_CONFIG_PATH += $(PWD)/vendor/github.com/limetext/rubex:$(GOPATH)/src/github.com/limetext/rubex
travis_test: test check_fmt check_license
