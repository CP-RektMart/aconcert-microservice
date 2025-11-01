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
