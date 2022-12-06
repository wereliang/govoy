FROM golang:1.18 as builder
WORKDIR /
COPY . .
ARG LDFLAGS
RUN echo $(LDFLAGS)
RUN go env -w GOPROXY=http://goproxy.cn,direct && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o http-server ./*.go

FROM centos
WORKDIR /
COPY --from=builder /http-server ./
ENTRYPOINT [ "/http-server", "8888" ]
