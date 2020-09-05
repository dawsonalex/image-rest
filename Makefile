build:
	go build

run: 
	go run main.go -dir=./sample_images

test: 
	go test ./...