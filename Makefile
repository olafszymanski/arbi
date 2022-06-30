include .env

run-dev:
	docker build \
	-t arbi \
	-f Dockerfile.dev \
	--build-arg GCP_PROJECT_ID=${GCP_PROJECT_ID} \
	--build-arg BINANCE_API_KEY=${BINANCE_API_KEY} \
	--build-arg BINANCE_SECRET_KEY=${BINANCE_SECRET_KEY} .
	docker run arbi