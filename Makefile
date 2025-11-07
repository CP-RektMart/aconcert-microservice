AUTH_DB=postgres://postgres:password@localhost:5431/auth?sslmode=disable
EVENT_DB=postgres://postgres:password@localhost:5433/event?sslmode=disable
RESERVATION_DB=postgres://postgres:password@localhost:5434/reservation?sslmode=disable

migrate-up:
	dbmate -d services/event/db/migrations -u ${EVENT_DB} up
	dbmate -d services/event/db/migrations -u ${EVENT_DB} up
	dbmate -d services/reservation/db/migrations -u ${RESERVATION_DB} up

migrate-down:
	dbmate -d services/auth/db/migrations -u ${AUTH_DB} down
	dbmate -d services/event/db/migrations -u ${EVENT_DB} down
	dbmate -d services/reservation/db/migrations -u ${RESERVATION_DB} down


sqlc:
	sqlc generate

compose-up:
	docker compose --env-file .env.port -f docker-compose.yaml up -d

compose-down:
	docker compose --env-file .env.port -f docker-compose.yaml down

protoc:
	protoc -I=pkg/proto \
		--go_out=pkg/proto --go_opt=paths=source_relative \
		--go-grpc_out=pkg/proto --go-grpc_opt=paths=source_relative \
		$(shell find pkg/proto -name "*.proto")

generate:
	go generate ./...
