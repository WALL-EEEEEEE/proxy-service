FROM golang:1.21.4-alpine3.18 AS build
WORKDIR .
ADD . /app/proxy/
RUN go build -o /go/bin/proxy /app/proxy/main.go
CMD ["/go/bin/proxy"]