default:
	$(MAKE) all

windows:
	mkdir -p dist/windows
	env GOOS=windows GOARCH=amd64 go build -o dist/windows/tokens2secrets.exe main.go

linux:
	mkdir -p dist/linux
	env GOOS=linux GOARCH=amd64 go build -o dist/linux/tokens2secrets main.go

darwin:
	mkdir -p dist/darwin
	env GOOS=darwin GOARCH=amd64 go build -o dist/darwin/tokens2secrets main.go

all:
	$(MAKE) windows
	$(MAKE) darwin
	$(MAKE) linux

dep:
	dep ensure -v