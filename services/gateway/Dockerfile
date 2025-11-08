FROM golang:1.25.0-alpine3.21

WORKDIR /app

COPY go.mod go.sum ./
COPY services/gateway/go.mod services/gateway/go.sum ./services/gateway/
COPY pkg ./pkg/

RUN cd services/gateway && go mod download

COPY services/gateway ./services/gateway/

RUN go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc4
RUN cd services/gateway && swag init -v3.1 -g cmd/main.go
RUN cd services/gateway && CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/main.go

EXPOSE 8080

CMD ["/app/app"]
