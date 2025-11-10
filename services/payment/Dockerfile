FROM golang:1.25.0-alpine3.21

WORKDIR /app

COPY go.mod go.sum ./
COPY services/payment/go.mod services/payment/go.sum ./services/payment/
COPY pkg ./pkg/

RUN cd services/payment && go mod download

COPY services/payment ./services/payment/

RUN cd services/payment && CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/main.go

EXPOSE 8080

CMD ["/app/app"]
