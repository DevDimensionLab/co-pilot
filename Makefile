.DEFAULT_GOAL := all
BUILD_DATE := `date +%Y-%m-%d\ %H:%M`
BUILD_TAG := `git describe --abbrev=0 --tags`
VERSION_FILE := cmd/version.go

genver:
	@rm -f $(VERSION_FILE)
	@echo "package cmd" > $(VERSION_FILE)
	@echo "const (" >> $(VERSION_FILE)
	@echo "  Version = \"$(BUILD_TAG)\"" >> $(VERSION_FILE)
	@echo "  BuildDate = \"$(BUILD_DATE)\"" >> $(VERSION_FILE)
	@echo ")" >> $(VERSION_FILE)

build:
	go build

docker-build:
	docker build --tag co-pilot:latest .

docker-run:
	docker run co-pilot $(ARGS)

install:
	go install

run:
	go run main.go

test:
	go test -v -cover ./...

lint:
	gofmt -w pkg
	gofmt -w cmd

release:
	goreleaser --rm-dist

upgrade:
	go get github.com/perottobc/mvn-pom-mutator


all: genver build install
