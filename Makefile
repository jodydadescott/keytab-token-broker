default:
	$(MAKE) windows

windows:
	mkdir -p dist
	env GOOS=windows GOARCH=amd64 go build -o dist/kbridge.exe main.go

linux:
	mkdir -p dist
	env GOOS=linux GOARCH=amd64 go build -o dist/kbridge-linux-amd64 main.go

darwin:
	mkdir -p dist
	env GOOS=darwin GOARCH=amd64 go build -o dist/kbridge-darwin-amd64 main.go

all:
	$(MAKE) windows
	$(MAKE) darwin
	$(MAKE) linux

dep:
	dep ensure -v
