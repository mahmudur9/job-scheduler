# Run parallelly
run:
	go run cmd/api/main.go & \
	go run cmd/scheduler/main.go & \
	go run cmd/worker/main.go & \
	wait

build:
	go build -o bin/api cmd/api/main.go
	go build -o bin/api cmd/scheduler/main.go
	go build -o bin/api cmd/worker/main.go

test:
	#go test ./...
	go test ./tests

fmt:
	go fmt ./...

docker-build:
	sudo docker compose build

docker-run:
	sudo docker compose up -d