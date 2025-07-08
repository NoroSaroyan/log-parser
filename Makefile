up:
	@docker compose -f deployments/docker-compose.yml --project-directory . up -d

docker-build:
	@docker compose -f deployments/docker-compose.yml --project-directory . build

