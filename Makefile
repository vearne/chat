VERSION :=v0.0.7

IMPORT_PATH = github.com/vearne/chat

BUILD_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date +%Y%m%d%H%M%S)
GITTAG = `git log -1 --pretty=format:"%H"`
LDFLAGS = -ldflags "-s -w -X ${IMPORT_PATH}/consts.GitTag=${GITTAG} -X ${IMPORT_PATH}/consts.BuildTime=${BUILD_TIME} -X ${IMPORT_PATH}/consts.Version=${VERSION}"

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o chat
