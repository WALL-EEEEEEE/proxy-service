# syntax = docker/dockerfile:1.3
FROM golang:1.21.4-alpine3.18 AS grpc-base


RUN  GOBIN=/go/bin go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@latest
RUN  GOBIN=/go/bin go install  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest 
RUN  GOBIN=/go/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN  GOBIN=/go/bin go install  google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN  GOBIN=/go/bin go install github.com/grpc-ecosystem/grpc-health-probe@latest

FROM golang:1.21.4-alpine3.18 AS app-build

COPY --from=0 /go/bin/* /go/bin/
COPY --from=bufbuild/buf /usr/local/bin/buf /usr/local/bin/buf


WORKDIR /app/proxy/gateway
COPY gateway/go.mod .
COPY gateway/go.sum .
#COPY proxy/gateway .
RUN go mod download && go mod tidy

WORKDIR /app/proxy/manager
COPY  manager/proto ./proto
COPY  manager/buf.gen.yaml .

RUN  cd /app/proxy/manager && buf generate proto

# run check_api job
WORKDIR /app/proxy
COPY . .
WORKDIR /app/proxy/gateway
RUN --mount=type=cache,mode=0755,target=/root/.cache/go-build --mount=type=cache,mode=0755,target=/root/go go build -o /go/bin/gateway cmd/http/main.go

FROM alpine
COPY --from=app-build /go/bin/gateway /app/proxy/gateway
CMD ["/app/proxy/gateway", "-m", "manager-server:8082", "-l", "DEBUG"]