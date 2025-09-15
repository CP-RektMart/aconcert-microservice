module github.com/cp-rektmart/aconcert-microservice/event

go 1.25.0

replace github.com/cp-rektmart/aconcert-microservice => ../../

require (
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250826171959-ef028d996bc1 // indirect
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
)

require (
	github.com/jackc/pgx/v5 v5.7.6
	google.golang.org/genproto v0.0.0-20250908214217-97024824d090
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/crypto v0.41.0 // indirect
)
