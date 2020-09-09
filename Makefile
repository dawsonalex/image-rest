LOGLEVEL ?= debug
WATCHDIR ?= ./sample_images

run: 
	GOOS=linux GOARCH=amd64 go build -o bin/imageservice
	bin/imageservice -dir=$(WATCHDIR) -l=$(LOGLEVEL) &

stop:
	pkill -f imageservice

clean:
	rm -rv bin/

test: 
	go test ./...