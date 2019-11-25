
build:
	$(GOROOT)/bin/go build -i -o ./dspotify ./src/main.go

build-docker: build
	docker build -f Docker/Dockerfile -t dspotify .
	rm dspotify