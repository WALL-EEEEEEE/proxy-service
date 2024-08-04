FROM golang:1.21.4-alpine3.18 AS grpc-base


RUN  GOBIN=/go/bin go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@latest
RUN  GOBIN=/go/bin go install  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest 
RUN  GOBIN=/go/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN  GOBIN=/go/bin go install  google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN  GOBIN=/go/bin go install github.com/grpc-ecosystem/grpc-health-probe@latest

FROM golang:1.21.4-alpine3.18 AS app-build

COPY --from=0 /go/bin/* /go/bin/
COPY --from=bufbuild/buf /usr/local/bin/buf /usr/local/bin/buf


WORKDIR /app/proxy/manager
COPY manager/go.mod .
COPY manager/go.sum .
#COPY proxy/manager .
RUN go mod download && go mod tidy


COPY manager/proto proto
COPY manager/buf.gen.yaml .
RUN  buf generate proto

WORKDIR /app/proxy/
COPY . .
WORKDIR /app/proxy/manager
RUN --mount=type=cache,mode=0755,target=/root/.cache/go-build --mount=type=cache,mode=0755,target=/root/go go build -o /go/bin/server ./cmd/server/main.go

FROM alpine
COPY --from=app-build /go/bin/server /app/proxy/server
COPY --from=app-build /go/bin/grpc-health-probe /app/proxy/grpc-health-probe
COPY manager/config.yml /app/proxy/config.yml
ENV LOGLEVEL=DEBUG
CMD ["/app/proxy/server", "-c", "/app/proxy/config.yml", "-l", "DEBUG"]