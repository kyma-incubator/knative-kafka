# ###############################################################################
#
#                       Knative-Kafka Common Makefile
#
# ###############################################################################

# Project Directories
BUILD_ROOT:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR:=$(BUILD_ROOT)/build

#
# Dep / Fmt / Vet / Lint / Test / Sonar
#

dep:
	@echo 'Ensuring Dependencies'
	cd $(BUILD_ROOT); dep ensure -v

fmt:
	@echo 'Formatting Go Source'
	go fmt ./...

vet:
	@echo 'Vetting Packages'
	go vet ./pkg/...

lint:
	@echo 'Linting Packages'
	golint ./pkg/...

test:
	@echo 'Testing Packages'
	cd $(BUILD_ROOT); mkdir -p $(BUILD_DIR);
	go test -v ./pkg/... -tags static -coverprofile ${BUILD_DIR}/coverage.out
	go tool cover -func=${BUILD_DIR}/coverage.out

.PHONY: dep fmt vet lint test

