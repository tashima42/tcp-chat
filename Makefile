REPO ?= tashima42
IMAGE = $(REPO)/tcp-chat:$(TAG)
BUILD_ACTION = --load
RUNNER := docker
IMAGE_BUILDER := $(RUNNER) buildx
MACHINE := tcp-chat
BUILDX_ARGS ?= --sbom=true --attest type=provenance,mode=max

buildx-machine: ## create rancher dockerbuildx machine targeting platform defined by DEFAULT_PLATFORMS.
	@docker buildx ls | grep $(MACHINE) || \
		docker buildx create --name=$(MACHINE)

push-image: buildx-machine ## build the container image targeting all platforms defined by TARGET_PLATFORMS and push to a registry.
	$(IMAGE_BUILDER) build -f Dockerfile \
		--builder $(MACHINE) $(BUILDX_ARGS) \
		-t "$(IMAGE)" --push .
	@echo "Pushed $(IMAGE)"
