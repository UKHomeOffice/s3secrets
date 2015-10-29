NAME=s3secrets
AUTHOR=UKHomeOffice
HARDWARE=$(shell uname -m)
PWD=$(shell pwd)
VERSION=$(shell awk '/const Version/ { print $$4 }' src/github.com/UKHomeOffice/s3secrets/main.go | sed 's/"//g')
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods -nilfunc -printf -rangeloops -shift -structtags

.PHONY: build docker-build deps vet format test clean travis release

default: docker-build

build:
	@echo "--> Performing a build"
	gb build all

docker-build:
	@echo "--> Performing a docker build"
	sudo docker run --rm -ti -v $(PWD):/go -w /go quay.io/ukhomeofficedigital/go-gb:1.0.0 gb build all

release:
	@$(MAKE) docker-build
	@echo "--> Performing a release"
	mkdir -p release/
	cp bin/${NAME} release/${NAME}_${VERSION}_linux_${HARDWARE}

clean:
	rm -rf ./secrets
	rm -rf ./bin

deps:
	@echo "--> Updating the dependencies"
	@which gb 2>/dev/null ; if [ $$? -eq 1 ]; then \
		go get github.com/constabulary/gb/...; \
	fi

lint:
	@echo "--> Running golint"
	@which golint 2>/dev/null ; if [ $$? -eq 1 ]; then \
		go get -u github.com/golang/lint/golint; \
	fi
	golint src/github.com/UKHomeOffice/s3secrets

vet:
	@echo "--> Running go tool vet"
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	go tool vet $(VETARGS) src/github.com/UKHomeOffice/s3secrets

format:
	@echo "--> Running go fmt"
	gofmt -d src/github.com/UKHomeOffice/s3secrets

test:
	@echo "--> Running go tests"
	go test -v
	@$(MAKE) vet

travis:
	@echo "--> Performing Unit tests"
	@$(MAKE) deps
	@$(MAKE) lint
	@$(MAKE) vet
	@$(MAKE) format
	@$(MAKE) build
