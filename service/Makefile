.PHONY:  test install build run lint

test:
	@scripts/test.sh local-dev

install:
	@scripts/install.sh

build:
	@scripts/build.sh local-dev

run:
	./build/assingment_service

lint:
	golangci-lint run
