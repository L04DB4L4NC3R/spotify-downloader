FROM golang
WORKDIR /app
COPY ./src/ytber .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ytber .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /app .
CMD ["./ytber"]
