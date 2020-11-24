FROM golang
WORKDIR /app
COPY ./src/scraper .

# Build protocol buffers
RUN apt-get update && apt-get install -y protobuf-compiler
RUN  go get google.golang.org/protobuf/cmd/protoc-gen-go \
         google.golang.org/grpc/cmd/protoc-gen-go-grpc
RUN protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
./proto/*.proto

# Compile
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scraper .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /app .
EXPOSE 3000
CMD ["./scraper"]
