BuildVersion = v1
Registry = docker.io
LDFlags = "-X 'main.BuildVersion=$(BuildVersion)'"
Image = $(Registry)/hunter2019/http-server:$(BuildVersion)

all:
	docker build --build-arg LDFLAGS=$(LDFlags) -t $(Image) .

push:
	docker push $(Image)