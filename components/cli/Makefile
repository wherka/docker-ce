<<<<<<< HEAD
#
# github.com/docker/cli
#
all: binary


_:=$(shell ./scripts/warn-outside-container $(MAKECMDGOALS))

.PHONY: clean
clean: ## remove build artifacts
	rm -rf ./build/* cli/winresources/rsrc_* ./man/man[1-9] docs/yaml/gen

.PHONY: test-unit
test-unit: ## run unit test
	./scripts/test/unit $(shell go list ./... | grep -vE '/vendor/|/e2e/')

.PHONY: test
test: test-unit ## run tests

.PHONY: test-coverage
test-coverage: ## run test coverage
	./scripts/test/unit-with-coverage $(shell go list ./... | grep -vE '/vendor/|/e2e/')

.PHONY: fmt
fmt:
	go list -f {{.Dir}} ./... | xargs gofmt -w -s -d

.PHONY: lint
lint: ## run all the lint tools
	gometalinter --config gometalinter.json ./...

.PHONY: binary
binary: ## build executable for Linux
	@echo "WARNING: binary creates a Linux executable. Use cross for macOS or Windows."
	./scripts/build/binary

.PHONY: cross
cross: ## build executable for macOS and Windows
	./scripts/build/cross

.PHONY: binary-windows
binary-windows: ## build executable for Windows
	./scripts/build/windows

.PHONY: binary-osx
binary-osx: ## build executable for macOS
	./scripts/build/osx

.PHONY: dynbinary
dynbinary: ## build dynamically linked binary
	./scripts/build/dynbinary

vendor: vendor.conf ## check that vendor matches vendor.conf
	rm -rf vendor
	bash -c 'vndr |& grep -v -i clone'
	scripts/validate/check-git-diff vendor

.PHONY: authors
authors: ## generate AUTHORS file from git history
	scripts/docs/generate-authors.sh

.PHONY: manpages
manpages: ## generate man pages from go source and markdown
	scripts/docs/generate-man.sh

.PHONY: yamldocs
yamldocs: ## generate documentation YAML files consumed by docs repo
	scripts/docs/generate-yaml.sh

.PHONY: shellcheck
shellcheck: ## run shellcheck validation
	scripts/validate/shellcheck

.PHONY: help
help: ## print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)


cli/compose/schema/bindata.go: cli/compose/schema/data/*.json
	go generate github.com/docker/cli/cli/compose/schema

compose-jsonschema: cli/compose/schema/bindata.go
	scripts/validate/check-git-diff cli/compose/schema/bindata.go

.PHONY: ci-validate
ci-validate:
	time make -B vendor
	time make -B compose-jsonschema
	time make manpages
	time make yamldocs
=======
CLI_DIR:=$(CURDIR)/components/cli
ENGINE_DIR:=$(CURDIR)/components/engine
PACKAGING_DIR:=$(CURDIR)/components/packaging
MOBY_COMPONENTS_SHA=ab7c118272b02d8672dc0255561d0c4015979780
MOBY_COMPONENTS_URL=https://raw.githubusercontent.com/docker/moby-extras/$(MOBY_COMPONENTS_SHA)/cmd/moby-components
MOBY_COMPONENTS=.helpers/moby-components-$(MOBY_COMPONENTS_SHA)
VERSION=$(shell cat VERSION)

.PHONY: help
help: ## show make targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf " \033[36m%-20s\033[0m  %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: test-integration-cli
test-integration-cli: $(CLI_DIR)/build/docker ## test integration of cli and engine
	$(MAKE) -C $(ENGINE_DIR) DOCKER_CLI_PATH=$< test-integration-cli

$(CLI_DIR)/build/docker:
	$(MAKE) -C $(CLI_DIR) -f docker.Makefile build

.PHONY: deb
deb: ## build deb packages
	$(MAKE) VERSION=$(VERSION) CLI_DIR=$(CLI_DIR) ENGINE_DIR=$(ENGINE_DIR) -C $(PACKAGING_DIR) deb

.PHONY: rpm
rpm: ## build rpm packages
	$(MAKE) VERSION=$(VERSION) CLI_DIR=$(CLI_DIR) ENGINE_DIR=$(ENGINE_DIR) -C $(PACKAGING_DIR) rpm

.PHONY: static
static: ## build static packages
	$(MAKE) VERSION=$(VERSION) CLI_DIR=$(CLI_DIR) ENGINE_DIR=$(ENGINE_DIR) -C $(PACKAGING_DIR) static

.PHONY: clean
clean: ## clean the build artifacts
	-$(MAKE) -C $(CLI_DIR) clean
	-$(MAKE) -C $(ENGINE_DIR) clean
	-$(MAKE) -C $(PACKAGING_DIR) clean

$(MOBY_COMPONENTS):
	mkdir -p .helpers
	curl -fsSL $(MOBY_COMPONENTS_URL) > $(MOBY_COMPONENTS)
	chmod +x $(MOBY_COMPONENTS)

.PHONY: update-components
update-components: update-components-cli update-components-engine update-components-packaging ## udpate components using moby extra tool

.PHONY: update-components-cli
update-components-cli: $(MOBY_COMPONENTS)
	$(MOBY_COMPONENTS) update cli

.PHONY: update-components-engine
update-components-engine: $(MOBY_COMPONENTS)
	$(MOBY_COMPONENTS) update engine

.PHONY: update-components-packaging
update-components-packaging: $(MOBY_COMPONENTS)
	$(MOBY_COMPONENTS) update packaging
>>>>>>> refs/components/cli/HEAD
