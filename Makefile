EVENT_DB = postgresql://postgres:password@localhost:5433/event-postgres?sslmode=disable

sqlc:
	sqlc generate

compose-up:
	docker compose -f docker-compose.yaml up -d

compose-down:
	docker compose -f docker-compose.yaml down

migrate-up:
	dbmate -d services/event/db/migrations -u ${EVENT_DB} up 

protoc:
	protoc -I=pkg/proto \
		--go_out=pkg/proto --go_opt=paths=source_relative \
		--go-grpc_out=pkg/proto --go-grpc_opt=paths=source_relative \
		$(shell find pkg/proto -name "*.proto")
