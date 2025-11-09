module github.com/cp-rektmart/aconcert-microservice/notification

go 1.25.0

replace github.com/cp-rektmart/aconcert-microservice => ../../

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/cp-rektmart/aconcert-microservice v0.0.0-20250917044643-2f895e009c55
	github.com/joho/godotenv v1.5.1
)

require github.com/streadway/amqp v1.1.0 // indirect
