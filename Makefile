GO111MODULE = on
CGO_ENABLED = 0
export GO111MODULE
export CGO_ENABLED

clean:
	rm -rf bin/
	rm -rf vendor/

build:
	mkdir -p bin
	cd bin && gox -arch="amd64" -os="darwin linux" -parallel=3 ../cmd/kubeconsole
	ls -al bin/

docker:
	docker build -t kubeconsole .
