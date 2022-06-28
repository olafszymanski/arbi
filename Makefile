run-dev:
	docker build . -t arbi -f Dockerfile.dev
	docker run arbi