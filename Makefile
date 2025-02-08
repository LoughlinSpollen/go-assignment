.PHONY: install-dev start-infra stop-infra test build install run stop

install-dev:
	@scripts/install-dev.sh

start-infra:
	brew services start rabbitmq

stop-infra:
	brew services stop rabbitmq

test:
	@service/scripts/test.sh local-dev
	@client/scripts/test.sh local-dev
	@shared_lib/scripts/test.sh local-dev

build:
	@client/scripts/build.sh local-dev
	@service/scripts/build.sh local-dev

install:
	@client/scripts/install.sh
	@service/scripts/install.sh

run:
	@scripts/run.sh local-dev

stop:
	@scripts/kill.sh 