BuildVersion = v1
Registry = docker.io
LDFlags = "-X 'main.BuildVersion=$(BuildVersion)'"
Image = $(Registry)/hunter2019/govoy:$(BuildVersion)

all:
	go build -ldflags $(LDFlags) -o govoy ./cmd/*.go

build:
	docker build --build-arg LDFLAGS=$(LDFlags) -t $(Image) .

push:
	docker push $(Image)