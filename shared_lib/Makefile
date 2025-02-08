.PHONY: watch test install

watch:
	@scripts/watch.sh

test:
	@scripts/config.sh local-dev
	@scripts/test.sh

install:
	@scripts/install.sh

lint: 
	golangci-lint run
