up:
	@docker compose -f deployments/docker-compose.yml --project-directory . up -d

down:
	@docker compose -f deployments/docker-compose.yml --project-directory . down

restart: down up

docker-build:
	@docker compose -f deployments/docker-compose.yml --project-directory . build

logs:
	@docker compose -f deployments/docker-compose.yml --project-directory . logs -f

ps:
	@docker compose -f deployments/docker-compose.yml --project-directory . ps

exec-app:
	@docker exec -it log-parser sh

exec-postgres:
	@docker exec -it my_postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)
