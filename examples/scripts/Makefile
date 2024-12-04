GO = $(shell which go)
DOCKER = $(shell which docker)

BUILD_FOLDER = ./build

.PHONY: build-linux
build-linux:
	@echo "+ $@"
	@rm -rf $(BUILD_FOLDER)
	@mkdir -p $(BUILD_FOLDER)
	@$(DOCKER) images | grep 'module-sdk-test' | awk '{print $$3}' | xargs $(DOCKER) rmi --force
	@$(DOCKER) build -f Dockerfile -t module-sdk-test --target final .
	@$(DOCKER) cp $(shell $(DOCKER) create module-sdk-test:latest ):/root/ $(BUILD_FOLDER)/

.PHONY: docker-clean
docker-clean:
	@echo "+ $@"
	@$(DOCKER) rm -vf $(shell $(DOCKER) ps -aq)