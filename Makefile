BINARY=gtm
VERSION=gtm-dev-$(shell date +'%Y.%m.%d-%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION}"
GIT2GO_VERSION=v27
GIT2GO_PATH=${GOPATH}/src/github.com/libgit2/git2go

build:
	go build --tags static  ${LDFLAGS} -o ${BINARY}

test:
	go test --tags static  $$(go list ./... | grep -v vendor)

vet:
	go vet $$(go list ./... | grep -v vendor)

fmt:
	go fmt $$(go list ./... | grep -v vendor)

install:
	go install --tags static ${LDFLAGS}

clean:
	go clean

git2go-install:
	[[ -d ${GIT2GO_PATH} ]] || git clone https://github.com/libgit2/git2go.git ${GIT2GO_PATH} && \
	cd ${GIT2GO_PATH} && \
	git pull && \
	git checkout -qf ${GIT2GO_VERSION} && \
	git submodule update --init

git2go-build:
	cd ${GIT2GO_PATH}/vendor/libgit2 && \
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

.PHONY: build test vet fmt install clean git2go-install git2go-build
