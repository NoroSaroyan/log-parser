help:
	@echo "Makefile commands:"
	@echo "  up           - Start all containers"
	@echo "  down         - Stop all containers"
	@echo "  restart      - Restart all containers"
	@echo "  docker-build - Build Docker images"
	@echo "  logs         - Follow logs"
	@echo "  ps           - Show running containers"
	@echo "  exec-app     - Exec shell into app container"
	@echo "  exec-postgres- Exec psql inside postgres container"
	@echo "  exec-godoc   - Exec shell into godoc container"
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
exec-godoc:
	@docker exec -it godoc sh
godoc:
	@docker compose -f deployments/docker-compose.yml --project-directory . up -d godoc
