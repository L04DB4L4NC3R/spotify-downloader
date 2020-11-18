.PHONY: build
build:
	cd ./src/scraper && go build -o ../../bin/scraper

.PHONY: run
run:
	make build
	./bin/scraper

.PHONY: watch
watch:
	reflex -s -r '\.go$$' make run
