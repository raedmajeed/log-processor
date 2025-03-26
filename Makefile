SERVICE_PATH=${PWD}/log-mainService/

GOFMT=gofmt -w
GOBUILD=go build -gcflags "-N -l" -v

ifeq ($(shell uname), Darwin)
    TARGET_OS := linux
    TARGET_ARCH := amd64
else
    TARGET_OS := linux
    TARGET_ARCH := amd64
endif

GOBUILD = env GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -gcflags "-N -l" -v

APP_NAME=log-main-service.bin
APP_DIR=main

APPTARGET=$(SERVICE_PATH)/bin/$(APP_NAME)

all:
	$(GOFMT) $(SERVICE_PATH)/$(APP_DIR)/*.go
	$(GOBUILD) -o $(APPTARGET) $(SERVICE_PATH)/$(APP_DIR)/*.go

docker:
	@echo "Building log-mainService Docker Image"
	docker build -f Dockerfile.log-mainService -t log-main-service:v1 ..
