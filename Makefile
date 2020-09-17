LOGLEVEL ?= debug
WATCHDIR ?= ./sample_images

build:
	go build -o bin/imageservice

run: build
	bin/imageservice -dir=$(WATCHDIR) -l=$(LOGLEVEL) &

stop:
	pkill -f imageservice

clean:
	rm -rv bin/

test: 
	go test ./...