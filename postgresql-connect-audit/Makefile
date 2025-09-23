.PHONY: build test clean deploy validate

build:
	sam build

validate:
	sam validate

test:
	cd postgresql-connect-audit && go test -v ./...

clean:
	sam delete

deploy:
	sam deploy --guided

