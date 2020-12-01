IMAGES = routing

TAG ?= latest
REPO_DIR := ${CURDIR}
REVISION := $(shell git -C $(REPO_DIR) rev-parse HEAD)

CLI_PATH = $(CURDIR)/cli

.PHONY: $(IMAGES)

all: build

build:
	cd $(CLI_PATH) && go build -o ../dist/cli

$(IMAGES):
	DOCKER_BUILDKIT=1 docker build $(REPO_DIR) \
		--file "$(REPO_DIR)/$@/docker/Dockerfile" \
		--build-arg GIT_SHA=$(REVISION) \
		--tag "eirini/$@:$(TAG)"

push-%:
	docker push eirini/$*:$(TAG)
