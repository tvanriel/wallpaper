.PHONY: docker
docker:
	docker buildx build --builder mybuilder --push --platform linux/amd64,linux/arm64 -f Dockerfile.prod -t mitaka8/wallpapers .

.PHONY: dev
dev:
	docker-compose up
