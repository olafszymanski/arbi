run_dev:
	docker build Dockerfile.dev -t arbi
	docker run arbi