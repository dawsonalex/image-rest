LOGLEVEL ?= debug
WATCHDIR ?= ./sample_images
PORT ?= 8080

build:
	go build -o bin/imageservice

run: build
	bin/imageservice -dir=$(WATCHDIR) -l=$(LOGLEVEL) -p=$(PORT) &

stop:
	pkill -f imageservice

clean:
	rm -rv bin/

test: 
	go test ./...