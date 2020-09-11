default:
	$(MAKE) windows

windows:
	mkdir -p dist/windows
	env GOOS=windows GOARCH=amd64 go build -o dist/windows/kbridge.exe main/main.go

linux:
	mkdir -p dist/linux
	env GOOS=linux GOARCH=amd64 go build -o dist/windows/kbridge main/main.go

darwin:
	mkdir -p dist/darwin
	env GOOS=darwin GOARCH=amd64 go build -o dist/darwin/kbridge main/main.go

all:
	$(MAKE) windows
	$(MAKE) darwin
	$(MAKE) linux

dep:
	dep ensure -v
