AUTH_DB = postgresql://root:password@localhost:5432/auth_db?sslmode=disable
EVENT_DB = postgresql://root:password@localhost:5433/event_db?sslmode=disable

sqlc:
	sqlc generate

compose-up:
	docker compose -f dockers/docker-compose.yaml up -d

compose-down:
	docker compose -f dockers/docker-compose.yaml down

migrate-up:
	dbmate -d services/event/db/migrations -u ${EVENT_DB} up 