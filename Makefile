# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

#Swagger ui
SWAGGERUIVERSION := 3.20.5

# Setup name variables for the package/tool
NAME := pairing-manager
PKG := github.com/sevren/test/$(NAME)
DOCKERTAG := pair-man

# Set any default go build tags
BUILDTAGS :=

# Set the build dir, where built cross-compiled binaries will be output
BUILDDIR := ${PREFIX}/cross

# Set our default go compiler
GO := go


# Populate version variables
# Add to compile time flags
VERSION := $(shell cat VERSION)
GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif
CTIMEVAR=-X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

# List the GOOS and GOARCH to build
GOOSARCHES = linux/amd64

all: clean build fmt lint test staticcheck vet install ## Runs a clean, build, fmt, lint, test, staticcheck, vet and install

.PHONY: build
build: $(NAME) ## Builds a dynamic executable or package

$(NAME): *.go VERSION
	@echo "+ $@"
	go build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(NAME) .

.PHONY: static
static: ## Builds a static executable
	@echo "+ $@"
	CGO_ENABLED=0 go build \
				-tags "$(BUILDTAGS) netgo static_build" \
				${GO_LDFLAGS_STATIC} -o $(NAME) .

.PHONY: docker
docker: static ## Builds a docker image
	@echo "+ $@"
	@docker build -t ${DOCKERTAG} --build-arg BINARY=$(NAME) .

.PHONY: fmt
fmt: ## Verifies all files have men `gofmt`ed
	@echo "+ $@"
	@gofmt -s -l . | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: lint
lint: ## Verifies `golint` passes
	@echo "+ $@"
	@golint ./... | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: test
test: ## Runs the go tests
	@echo "+ $@"
	@go test -v -tags "$(BUILDTAGS) cgo" $(shell go list ./... | grep -v vendor)

.PHONY: vet
vet: ## Verifies `go vet` passes
	@echo "+ $@"
	@go vet $(shell go list ./... | grep -v vendor) | grep -v '.pb.go:' | tee /dev/stderr

.PHONY: staticcheck
staticcheck: ## Verifies `staticcheck` passes
	@echo "+ $@"
	@staticcheck $(shell go list ./... | grep -v vendor) | grep -v '.pb.go:' | tee /dev/stderr

.PHONY: install
install: ## Installs the executable or package
	@echo "+ $@"
	go install -a -tags "$(BUILDTAGS)" ${GO_LDFLAGS} .

define buildpretty
mkdir -p $(BUILDDIR)/$(1)/$(2);
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=1 go build \
	 -o $(BUILDDIR)/$(1)/$(2)/$(NAME) \
	 -a -tags "$(BUILDTAGS) cgo static_build netgo" \
	 -installsuffix netgo ${GO_LDFLAGS_STATIC} .;
md5sum $(BUILDDIR)/$(1)/$(2)/$(NAME) > $(BUILDDIR)/$(1)/$(2)/$(NAME).md5;
sha256sum $(BUILDDIR)/$(1)/$(2)/$(NAME) > $(BUILDDIR)/$(1)/$(2)/$(NAME).sha256;
endef

.PHONY: cross
cross: *.go VERSION ## Builds the cross-compiled binaries, creating a clean directory structure (eg. GOOS/GOARCH/binary)
	@echo "+ $@"
	$(foreach GOOSARCH,$(GOOSARCHES), $(call buildpretty,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH))))

define buildrelease
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=1 go build \
	 -o $(BUILDDIR)/$(NAME)-$(1)-$(2) \
	 -a -tags "$(BUILDTAGS) cgo static_build netgo" \
	 -installsuffix netgo ${GO_LDFLAGS_STATIC} .;
md5sum $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).md5;
sha256sum $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).sha256;
endef

.PHONY: release
release: *.go VERSION ## Builds the cross-compiled binaries, naming them in such a way for release (eg. binary-GOOS-GOARCH)
	@echo "+ $@"
	$(foreach GOOSARCH,$(GOOSARCHES), $(call buildrelease,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH))))


.PHONY: clean
clean: ## Cleanup any build binaries or packages
	@echo "+ $@"
	$(RM) $(NAME)
	$(RM) -r $(BUILDDIR)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: swaggerui-download
swaggerui-download: ## Downloads swaggerui release
	@curl -L https://github.com/swagger-api/swagger-ui/archive/v$(SWAGGERUIVERSION).zip --output /tmp/v$(SWAGGERUIVERSION).zip
	@unzip -o -d /tmp/ /tmp/v$(SWAGGERUIVERSION).zip
	@mkdir -p swaggerui

.PHONY: swaggerui
swaggerui: swaggerui-download ## Generates static files for swaggerui
	@GO111MODULE=off $(GO) get github.com/rakyll/statik
	cp openapi.yml /tmp/swagger-ui-$(SWAGGERUIVERSION)/dist/openapi.yml
	# Change the swaggerui config file to point to our openapi.yml definition file and also set the redirect url
	sed -i -e 's|url: "https://petstore.swagger.io/v2/swagger.json",|url: "./openapi.yml",\n\toauth2RedirectUrl: window.location.origin + window.location.pathname + "oauth2-redirect.html",|g' /tmp/swagger-ui-3.20.5/dist/index.html
	statik -f -c="Static assets for swaggerui v$(SWAGGERUIVERSION)" -src=/tmp/swagger-ui-$(SWAGGERUIVERSION)/dist -dest=. -p=swaggerui
