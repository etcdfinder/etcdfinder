docker-build-test:
	docker buildx build --platform linux/amd64,linux/arm64 -t etcdfinder/etcdfinder:test .

docker-build:
	docker buildx build --platform linux/amd64,linux/arm64 -t etcdfinder/etcdfinder:latest .