.PHONY: build test clean deploy validate

build:
	sam build

validate:
	sam validate

test:
	cd db-connect-audit && go test -v ./...

clean:
	sam delete

deploy:
	sam deploy --guided

deploy-prod:
	sam deploy --parameter-overrides \
		SlackWebhookUrl=$(SLACK_WEBHOOK_URL) \
		AuroraClusterIdentifier=$(AURORA_CLUSTER_ID) \
		CloudWatchLogGroupName=$(LOG_GROUP_NAME)
