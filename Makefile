.PHONY: build
build:
	cd ./src/scraper && go build -o ../../bin/scraper
	cd ./src/ytber && go build -o ../../bin/ytber

.PHONY: run
run:
	make build
	./bin/scraper

.PHONY: watch
watch:
	reflex -s -r '\.go$$' make run

.PHONY: build-proto
build-proto:
	cd src/scraper/ && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
./proto/*.proto

	cd src/ytber/ && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=require_unimplemented_servers=false:. --go-grpc_opt=paths=source_relative \
./proto/*.proto
