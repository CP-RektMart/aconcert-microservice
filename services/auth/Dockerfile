FROM golang:1.25.0-alpine3.21

WORKDIR /app

COPY go.mod go.sum ./
COPY services/auth/go.mod services/auth/go.sum ./services/auth/
COPY pkg ./pkg/

RUN cd services/auth && go mod download

COPY services/auth ./services/auth/

RUN cd services/auth && CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/main.go

EXPOSE 8080

CMD ["/app/app"]
