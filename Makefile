LOGLEVEL ?= debug
WATCHDIR ?= ./sample_images
PORT ?= 3000
RELEASE_PLATFORMS := linux windows darwin# release platforms to build for
os = $(word 1, $@)
EXECUTABLE=imageservice
VERSION=$(shell git describe --tags --always)
RELEASE_DIR?=release

.PHONY: $(RELEASE_PLATFORMS)
$(RELEASE_PLATFORMS): release_dir
	GOOS=$(os) GOARCH=amd64 go build -o $(RELEASE_DIR)/$(VERSION)-$(EXECUTABLE)-$(os)-amd64
	GOOS=$(os) GOARCH=386 go build -o $(RELEASE_DIR)/$(VERSION)-$(EXECUTABLE)-$(os)-386

release_dir:
	mkdir -p $(RELEASE_DIR)

.PHONY: build_release
build_release: $(RELEASE_PLATFORMS)

build:
	go build -o bin/imageservice

run: build
	bin/imageservice -dir=$(WATCHDIR) -l=$(LOGLEVEL) -p=$(PORT) &

run_attached: build
	bin/imageservice -dir=$(WATCHDIR) -l=$(LOGLEVEL) -p=$(PORT)

stop:
	pkill -f imageservice

clean:
	rm -rfv bin/
	rm -rfv $(RELEASE_DIR)/

test: 
	go test ./...