VERSION :=v0.1.1

RELEASE_DIR = dist
IMPORT_PATH = github.com/vearne/chat

BUILD_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date +%Y%m%d%H%M%S)
GITTAG = `git log -1 --pretty=format:"%H"`
LDFLAGS = -ldflags "-s -w -X ${IMPORT_PATH}/consts.GitTag=${GITTAG} -X ${IMPORT_PATH}/consts.BuildTime=${BUILD_TIME} -X ${IMPORT_PATH}/consts.Version=${VERSION}"

TAG = ${VERSION}-${BUILD_TIME}-${BUILD_COMMIT}
IMAGE_BROKER = woshiaotian/chat-broker:${TAG}
IMAGE_LOGIC = woshiaotian/chat-logic:${TAG}


.PHONY: clean
clean: ## Remove release binaries
	rm -rf ${RELEASE_DIR}

build-dirs: clean
	mkdir -p ${RELEASE_DIR}

.PHONY: build
build: build-dirs
	env GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${RELEASE_DIR}/chat-broker ./cmd/broker
	env GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${RELEASE_DIR}/chat-logic ./cmd/logic
	chmod +x ${RELEASE_DIR}/*

.PHONY: image
image: build
	# broker
	docker build -f ./dockerfile/Dockerfile.broker --rm --no-cache -t ${IMAGE_BROKER} .
	docker push ${IMAGE_BROKER}
	# logic
	docker build -f ./dockerfile/Dockerfile.logic --rm --no-cache -t ${IMAGE_LOGIC} .
	docker push ${IMAGE_LOGIC}

