.PHONY: build up down restart logs logs-bot shell clean

# Build the Docker image
build:
	docker compose build

# Start the container (with build if needed)
up:
	docker compose up -d --build

# Stop and remove the container
down:
	docker compose down

# Restart the container
restart:
	docker compose down && docker compose up -d

# Follow all container logs
logs:
	docker compose logs -f

# Follow only bot logs (filter out KasmVNC noise)
logs-bot:
	docker compose logs -f 2>&1 | grep -E "^obsidian-pa.*202[0-9]/"

# Open a shell in the container
shell:
	docker compose exec obsidian-brain bash

# Show container stats
stats:
	docker stats --no-stream

# Remove container and image (keeps data)
clean:
	docker compose down --rmi local

# Remove everything including data (DESTRUCTIVE)
clean-all:
	docker compose down --rmi local -v
	rm -rf obsidian_data

# Local development
dev-build:
	go build -mod=vendor -o bot .

dev-test:
	go test -mod=vendor ./...
