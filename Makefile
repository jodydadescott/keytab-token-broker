default:
	$(MAKE) all

windows:
	$(MAKE) dep
	mkdir -p dist
	env GOOS=windows GOARCH=amd64 go build -o dist/ktbserver-windows-amd64.exe main.go

linux:
	$(MAKE) dep
	mkdir -p dist
	env GOOS=linux GOARCH=amd64 go build -o dist/ktbserver-linux-amd64 main.go

darwin:
	$(MAKE) dep
	mkdir -p dist
	env GOOS=darwin GOARCH=amd64 go build -o dist/ktbserver-darwin-amd64 main.go

all:
	$(MAKE) windows
	$(MAKE) darwin
	$(MAKE) linux

dep:
	dep ensure -v