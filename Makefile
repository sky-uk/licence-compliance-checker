.DEFAULT_GOAL := build
.PHONY: build fmt check vet lint test

PROJECT_DIR := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
BUILD_DIR := $(PROJECT_DIR)/build

junit_report_dir = $(BUILD_DIR)/junit-reports
pkgs := $(shell go list ./... | grep -v /vendor/ )

setup :
	@echo "== setup"
	go get -v golang.org/x/lint/golint golang.org/x/tools/cmd/goimports github.com/golang/dep/cmd/dep gopkg.in/src-d/go-license-detector.v2/...
	go get -v github.com/onsi/ginkgo/ginkgo && cd $$GOPATH/src/github.com/onsi/ginkgo && git checkout 'v1.6.0' && go install github.com/onsi/ginkgo/ginkgo
	dep ensure -v

build: ensure-build-dir-exists
	@echo "== build"
	go build -o $(BUILD_DIR)/bin/licence-compliance-checker -v github.com/sky-uk/licence-compliance-checker/cmd
	go test -run xxxxx $(pkgs)  # build the test code but don't run any tests yet

fmt:
	go fmt ./...

check: fmt vet lint test

vet:
	go vet $(pkgs)

lint:
	for pkg in $(pkgs); do \
		golint -set_exit_status $$pkg || exit 1; \
	done;

ensure-build-dir-exists:
	mkdir -p $(BUILD_DIR)

ensure-test-report-dir-exists: ensure-build-dir-exists
	mkdir -p $(junit_report_dir)

test: ensure-test-report-dir-exists
	@echo "== test"
	ginkgo -r --v --progress pkg cmd test/e2e -- -junit-report-dir $(junit_report_dir)

install: build
	@echo "== install"
	cp -v $(BUILD_DIR)/bin/licence-compliance-checker $(shell go env GOPATH)/bin/licence-compliance-checker

clean:
	rm -rfv $(BUILD_DIR)
