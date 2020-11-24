FROM golang
WORKDIR /app
COPY ./src/scraper .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scraper .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /app .
CMD ["./scraper"]
