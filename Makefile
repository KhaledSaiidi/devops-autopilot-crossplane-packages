SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

REGISTRY ?= ghcr.io
OWNER ?=
REPOSITORY ?=
TYPE ?=
NAME ?=
VERSION ?=
DIST_DIR ?= dist
GO_VERSION ?= 1

ifeq ($(TYPE),function)
ARTIFACT_NAME := function-$(NAME)
SOURCE_DIR := functions/$(NAME)
PACKAGE_ROOT := $(SOURCE_DIR)/package
LOCAL_RUNTIME_IMAGE := $(ARTIFACT_NAME):$(VERSION)-runtime
else ifeq ($(TYPE),pkg)
ARTIFACT_NAME := pkgs-$(NAME)
SOURCE_DIR := pkgs/$(NAME)
PACKAGE_ROOT := $(SOURCE_DIR)
LOCAL_RUNTIME_IMAGE :=
else
ARTIFACT_NAME :=
SOURCE_DIR :=
PACKAGE_ROOT :=
LOCAL_RUNTIME_IMAGE :=
endif

OWNER_LC := $(shell printf '%s' '$(OWNER)' | tr '[:upper:]' '[:lower:]')
REPOSITORY_LC := $(shell printf '%s' '$(REPOSITORY)' | tr '[:upper:]' '[:lower:]')
PACKAGE_FILE := $(DIST_DIR)/$(ARTIFACT_NAME)-$(VERSION).xpkg
ARTIFACT_REPO := $(REGISTRY)/$(OWNER_LC)/$(REPOSITORY_LC)/$(ARTIFACT_NAME)
ARTIFACT_REF := $(ARTIFACT_REPO):$(VERSION)

.PHONY: help
help:
	@echo "Crossplane OCI artifact tooling"
	@echo ""
	@echo "Targets:"
	@echo "  list-artifacts   List valid packages and functions in this repo"
	@echo "  metadata         Print resolved metadata for TYPE/NAME"
	@echo "  build            Build one OCI package file into dist/"
	@echo "  publish          Build and push one OCI package to GHCR"
	@echo "  clean-dist       Remove generated package files"
	@echo ""
	@echo "Required variables for build/publish:"
	@echo "  TYPE=function|pkg"
	@echo "  NAME=<directory-name>"
	@echo "  VERSION=vX.Y.Z"
	@echo ""
	@echo "Required variables for publish:"
	@echo "  OWNER=<github-owner>"
	@echo "  REPOSITORY=<github-repository-name>"
	@echo ""
	@echo "Examples:"
	@echo "  make build TYPE=pkg NAME=bucket VERSION=v1.0.0"
	@echo "  make publish TYPE=function NAME=bucket-name-normalizer VERSION=v1.0.0 OWNER=KhaledSaiidi REPOSITORY=devops-autopilot-crossplane-packages"

.PHONY: list-artifacts
list-artifacts:
	@if [ -d pkgs ]; then \
		find pkgs -mindepth 1 -maxdepth 1 -type d | sort | while read -r dir; do \
			if [ -f "$$dir/crossplane.yaml" ]; then \
				name="$$(basename "$$dir")"; \
				printf 'pkg\t%s\tpkgs-%s\t%s\n' "$$name" "$$name" "$$dir"; \
			fi; \
		done; \
	fi
	@if [ -d functions ]; then \
		find functions -mindepth 1 -maxdepth 1 -type d | sort | while read -r dir; do \
			if [ -f "$$dir/package/crossplane.yaml" ] && [ -f "$$dir/Dockerfile" ] && [ -f "$$dir/go.mod" ]; then \
				name="$$(basename "$$dir")"; \
				printf 'function\t%s\tfunction-%s\t%s\n' "$$name" "$$name" "$$dir"; \
			fi; \
		done; \
	fi

.PHONY: validate-inputs
validate-inputs:
	@[ -n "$(TYPE)" ] || { echo "TYPE is required"; exit 1; }
	@[ -n "$(NAME)" ] || { echo "NAME is required"; exit 1; }
	@case "$(TYPE)" in \
		function|pkg) ;; \
		*) echo "TYPE must be either 'function' or 'pkg'"; exit 1 ;; \
	esac
	@if [ "$(TYPE)" = "pkg" ]; then \
		[ -d "$(SOURCE_DIR)" ] || { echo "Package directory '$(SOURCE_DIR)' not found"; exit 1; }; \
		[ -f "$(SOURCE_DIR)/crossplane.yaml" ] || { echo "Package metadata '$(SOURCE_DIR)/crossplane.yaml' not found"; exit 1; }; \
	else \
		[ -d "$(SOURCE_DIR)" ] || { echo "Function directory '$(SOURCE_DIR)' not found"; exit 1; }; \
		[ -f "$(SOURCE_DIR)/Dockerfile" ] || { echo "Function Dockerfile '$(SOURCE_DIR)/Dockerfile' not found"; exit 1; }; \
		[ -f "$(SOURCE_DIR)/go.mod" ] || { echo "Function module file '$(SOURCE_DIR)/go.mod' not found"; exit 1; }; \
		[ -f "$(PACKAGE_ROOT)/crossplane.yaml" ] || { echo "Function package metadata '$(PACKAGE_ROOT)/crossplane.yaml' not found"; exit 1; }; \
	fi

.PHONY: validate-version
validate-version:
	@[ -n "$(VERSION)" ] || { echo "VERSION is required"; exit 1; }
	@[[ "$(VERSION)" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$$ ]] || { echo "VERSION must use semantic version format like v1.0.0"; exit 1; }

.PHONY: validate-publish
validate-publish:
	@[ -n "$(OWNER)" ] || { echo "OWNER is required for publish"; exit 1; }
	@[ -n "$(REPOSITORY)" ] || { echo "REPOSITORY is required for publish"; exit 1; }

.PHONY: metadata
metadata: validate-inputs
	@printf 'TYPE=%s\n' "$(TYPE)"
	@printf 'NAME=%s\n' "$(NAME)"
	@printf 'ARTIFACT_NAME=%s\n' "$(ARTIFACT_NAME)"
	@printf 'SOURCE_DIR=%s\n' "$(SOURCE_DIR)"
	@printf 'PACKAGE_ROOT=%s\n' "$(PACKAGE_ROOT)"
	@printf 'PACKAGE_FILE=%s\n' "$(PACKAGE_FILE)"
	@printf 'ARTIFACT_REPO=%s\n' "$(ARTIFACT_REPO)"
	@printf 'ARTIFACT_REF=%s\n' "$(ARTIFACT_REF)"

.PHONY: build
build: validate-inputs validate-version
	@case "$(TYPE)" in \
		pkg) \
			$(MAKE) build-pkg TYPE="$(TYPE)" NAME="$(NAME)" VERSION="$(VERSION)" DIST_DIR="$(DIST_DIR)" ;; \
		function) \
			$(MAKE) build-function TYPE="$(TYPE)" NAME="$(NAME)" VERSION="$(VERSION)" DIST_DIR="$(DIST_DIR)" GO_VERSION="$(GO_VERSION)" ;; \
	esac

.PHONY: build-pkg
build-pkg: validate-inputs validate-version
	@[ "$(TYPE)" = "pkg" ] || { echo "build-pkg requires TYPE=pkg"; exit 1; }
	@mkdir -p "$(DIST_DIR)"
	@examples_arg=""; \
	if [ -d "$(SOURCE_DIR)/examples" ]; then \
		examples_arg="--examples-root=$(SOURCE_DIR)/examples"; \
	fi; \
	echo "Building package $(ARTIFACT_NAME) from $(SOURCE_DIR)"; \
	crossplane xpkg build \
		--package-root="$(SOURCE_DIR)" \
		--package-file="$(PACKAGE_FILE)" \
		$$examples_arg
	@echo "Built $(PACKAGE_FILE)"

.PHONY: build-function
build-function: validate-inputs validate-version
	@[ "$(TYPE)" = "function" ] || { echo "build-function requires TYPE=function"; exit 1; }
	@mkdir -p "$(DIST_DIR)"
	@echo "Building runtime image $(LOCAL_RUNTIME_IMAGE) from $(SOURCE_DIR)"
	@docker build \
		--build-arg GO_VERSION="$(GO_VERSION)" \
		--tag "$(LOCAL_RUNTIME_IMAGE)" \
		"$(SOURCE_DIR)"
	@echo "Building function package $(ARTIFACT_NAME) from $(PACKAGE_ROOT)"
	@crossplane xpkg build \
		--package-root="$(PACKAGE_ROOT)" \
		--package-file="$(PACKAGE_FILE)" \
		--embed-runtime-image="$(LOCAL_RUNTIME_IMAGE)"
	@echo "Built $(PACKAGE_FILE)"

.PHONY: publish
publish: validate-inputs validate-version validate-publish build
	@echo "Publishing $(ARTIFACT_REF)"
	@crossplane xpkg push -f "$(PACKAGE_FILE)" "$(ARTIFACT_REF)"
	@echo "Published $(ARTIFACT_REF)"

.PHONY: clean-dist
clean-dist:
	@rm -rf "$(DIST_DIR)"
	@echo "Removed $(DIST_DIR)"
