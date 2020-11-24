FROM golang
WORKDIR /app
COPY ./src/ytber .

RUN apt-get update && apt-get install -y protobuf-compiler
RUN  go get google.golang.org/protobuf/cmd/protoc-gen-go \
         google.golang.org/grpc/cmd/protoc-gen-go-grpc
RUN protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=require_unimplemented_servers=false:. --go-grpc_opt=paths=source_relative \
./proto/*.proto

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ytber .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /app .
CMD ["./ytber"]
