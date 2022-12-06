FROM golang:1.18 as builder
WORKDIR /govoy/
COPY . .
ARG LDFLAGS
RUN echo $(LDFLAGS)
RUN go env -w GOPROXY=http://goproxy.cn,direct && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o govoy ./cmd/*.go

# FROM centos:centos7
FROM hunter2019/centos:7
WORKDIR /govoy/
COPY --from=builder /govoy/govoy ./bin/
COPY --from=builder /govoy/tools/iptables.sh ./bin/
COPY --from=builder /govoy/config/envoy0.yaml ./config/
RUN groupadd -g 1337 istio-proxy && useradd istio-proxy -u 1337 -g 1337
USER istio-proxy
ENTRYPOINT [ "./bin/govoy", "-c", "./config/envoy0.yaml" ]
