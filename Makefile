build:
	mkdir -p bin
	cd bin && gox -arch="amd64" -os="windows darwin linux" ../cmd/kubeconsole
