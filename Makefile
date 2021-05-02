.PHONY: build run kill watch build-proto cli
build:
	cd ./src/scraper && go build -o ../../bin/scraper
	cd ./src/ytber && go build -o ../../bin/ytber

run:
	make kill || echo Starting new process
	make build
	./bin/scraper &
	./bin/ytber &

kill:
	pkill scraper
	pkill ytber

watch:
	reflex -s -r '\.go$$' make run

build-proto:
	cd src/scraper/ && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
./proto/*.proto

	cd src/ytber/ && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=require_unimplemented_servers=false:. --go-grpc_opt=paths=source_relative \
./proto/*.proto

cli: build-proto
	cd src/scraper/cli && GOOS=linux GOARCH=amd64 go build -o sdl main.go && mv sdl ../../../sdl-linux-amd64
	cd src/scraper/cli && GOOS=linux GOARCH=arm go build -o sdl main.go && mv sdl ../../../sdl-linux-arm
	cd src/scraper/cli && GOOS=darwin GOARCH=amd64 go build -o sdl main.go && mv sdl ../../../sdl-mac-amd64
	cd src/scraper/cli && GOOS=darwin GOARCH=arm64 go build -o sdl main.go && mv sdl ../../../sdl-mac-arm64
