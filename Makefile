run_dev:
	docker build . -t arbi -f Dockerfile.dev
	docker run arbi