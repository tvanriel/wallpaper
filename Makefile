.PHONY: docker
docker:
	docker build -f Dockerfile.prod -t tvanriel/wallpapers .

.PHONY: dev
dev:
	docker-compose up
