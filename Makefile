GO111MODULE = on
CGO_ENABLED = 0
export GO111MODULE
export CGO_ENABLED

build:
	mkdir -p bin
	cd bin && gox -arch="amd64" -os="windows darwin linux" -parallel=3 ../cmd/kubeconsole
