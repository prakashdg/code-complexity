##################################################
# Default variables
##################################################

# Version string
include mk/version.mk

# Docker OPTIONS
DOCKER_IMG              ?= ci-tools
DOCKER_TAG              ?= latest
DOCKER_CACHE_TAG        ?= latest
DOCKER_CMD              ?= /bin/bash
DOCKER_REGISTRY         ?= 123456.dkr.ecr.us-west-2.amazonaws.com
DOCKER_REGISTRY_CACHE   ?= registry.eng.test.net
DOCKER_INTERACTIVE      ?= -it

# Fixed variables
DOCKER_BUILDKIT := 1
BUILDKIT_INLINE_CACHE := 1

##################################################
# Help targets (keep up to date with changes)
##################################################

help-all: help terraform-help

help:
	@echo ""
	@echo "Welcome!!! You are building version: $(VERSION)"
	@echo ""
	@echo "Specify a TARGET to run and any environment OPTIONS"
	@echo ""
	@echo "If you are running docker-* commands, you can set your"
	@echo "environment OPTIONS in .docker.env file at the root of"
	@echo "this repo"
	@echo ""
	@echo "OPTIONS (affected target in parenthesis) [default in brackets]"
	@echo ""
	@echo "  DOCKER_IMG         - (docker-*)" \
	"Name of docker image to build [$(DOCKER_IMG)]"
	@echo "  DOCKER_REGISTRY    - (docker-*)" \
	"Name of docker registry [$(DOCKER_REGISTRY)]"
	@echo "  DOCKER_TAG         - (docker-*)" \
	"Publishing tag for docker image [$(DOCKER_TAG)]"
	@echo "  JENKINS_USER       - (jenkins-jobs)" \
	@echo ""

##################################################
# Artifact producing targets
##################################################
.PHONY: version clean

version:
	@echo $(VERSION)

clean:
	@echo "=> Not Implemented"
	
.docker.env:
	touch .docker.env

docker-login:
	aws ecr get-login-password --region us-west-2 \
	| docker login --username AWS --password-stdin $(DOCKER_REGISTRY)

docker-clean:
	@echo "=> Cleaning docker images, containers and volumes"
	docker images | grep $(DOCKER_IMG) | awk '{print $$1":"$$2}' | xargs docker rmi || :
	docker system prune -f

docker-build: .docker.env
	@echo "=> Building docker dev environment"
	docker build \
	 --build-arg BUILDKIT_INLINE_CACHE=$(BUILDKIT_INLINE_CACHE) \
	 -t $(DOCKER_IMG):$(DOCKER_TAG) .

docker-dev: docker-build
	@echo "=> Launching interactive docker dev environment"
	docker run --rm -u `id -u`:`id -g` --network=host \
	--env-file .docker.env $(DOCKER_INTERACTIVE) -v $(CURDIR):/src \
	$(DOCKER_IMG):$(DOCKER_TAG) $(DOCKER_CMD)

docker-publish: docker-build
	@echo "=> Publishing docker image $(DOCKER_IMG):$(VERSION)"
	docker tag $(DOCKER_IMG):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMG):$(VERSION)
	docker tag $(DOCKER_IMG):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMG):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMG):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMG):$(DOCKER_TAG)

docker-run-% docker-make-%: docker-build
	@echo "=> Running make $* in brkt-infra-dev container"
	docker run --rm --env-file .docker.env $(DOCKER_IMG) make $*
