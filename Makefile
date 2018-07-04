BINARY         = bin/gtm
VERSION        = 0.0.0-dev
COMMIT         = $(shell git show -s --format='%h' HEAD)
LDFLAGS        = -ldflags "-X main.Version=$(VERSION)-$(COMMIT)"
GIT2GO_VERSION = v27
GIT2GO_PATH    = $(GOPATH)/src/github.com/libgit2/git2go
LIBGIT2_PATH   = $(GIT2GO_PATH)/vendor/libgit2
PKGS           = $(shell go list ./... | grep -v vendor)
BUILD_TAGS     = static

build:
	go build --tags '$(BUILD_TAGS)' $(LDFLAGS) -o $(BINARY)

debug: BUILD_TAGS += debug
debug: build

profile: BUILD_TAGS += profile
profile: build

debug-profile: BUILD_TAGS += debug profile
debug-profile: build

test:
	@go test $(TEST_OPTIONS) --tags '$(BUILD_TAGS)' $(PKGS) | grep --colour -E "FAIL|$$"

test-verbose: TEST_OPTIONS += -v
test-verbose: test

lint:
	-@$(call color_echo, 4, "\nGo Vet"); \
		go vet --all --tags '$(BUILD_TAGS)' $(PKGS)
	-@$(call color_echo, 4, "\nError Check"); \
		errcheck -ignoretests -tags '$(BUILD_TAGS)' $(PKGS)
	-@$(call color_echo, 4, "\nIneffectual Assign"); \
		ineffassign ./
	-@$(call color_echo, 4, "\nStatic Check"); \
		staticcheck --tests=false --tags '$(BUILD_TAGS)' $(PKGS)
	-@$(call color_echo, 4, "\nGo Simple"); \
		gosimple --tests=false --tags '$(BUILD_TAGS)' $(PKGS)
	-@$(call color_echo, 4, "\nUnused"); \
		unused --tests=false --tags '$(BUILD_TAGS)' $(PKGS)
	-@$(call color_echo, 4, "\nGo Lint"); \
		golint $(PKGS)
	-@$(call color_echo, 4, "\nGo Format"); \
		go fmt $(PKGS)
	-@$(call color_echo, 4, "\nLicense Check"); \
		ag --go -L license . |grep -v vendor/

install:
	go install --tags '$(BUILD_TAGS)' $(LDFLAGS)

clean:
	go clean
	rm bin/*

git2go-install:
	[[ -d $(GIT2GO_PATH) ]] || git clone https://github.com/libgit2/git2go.git $(GIT2GO_PATH) && \
	cd ${GIT2GO_PATH} && \
	git pull && \
	git checkout -qf $(GIT2GO_VERSION) && \
	git submodule update --init

git2go: git2go-install
	cd $(LIBGIT2_PATH) && \
	mkdir -p install/lib && \
	mkdir -p build && \
	cd build && \
	cmake -DTHREADSAFE=ON \
		  -DBUILD_CLAR=OFF \
		  -DBUILD_SHARED_LIBS=OFF \
		  -DCMAKE_C_FLAGS=-fPIC \
		  -DUSE_SSH=OFF \
		  -DCURL=OFF \
		  -DUSE_HTTPS=OFF \
		  -DUSE_BUNDLED_ZLIB=ON \
		  -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
		  -DCMAKE_INSTALL_PREFIX=../install \
		  .. && \
	cmake --build .

git2go-clean:
	[[ -d $(GIT2GO_PATH) ]] && rm -rf $(GIT2GO_PATH)

define color_echo
      @tput setaf $1
      @echo $2
      @tput sgr0
endef

.PHONY: build test vet fmt install clean git2go-install git2go-build all-tags profile debug
