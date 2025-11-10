FROM golang:1.25.0-alpine3.21

WORKDIR /app

COPY go.mod go.sum ./
COPY services/realtime/go.mod services/realtime/go.sum ./services/realtime/
COPY pkg ./pkg/

RUN cd services/realtime && go mod download

COPY services/realtime ./services/realtime/

RUN cd services/realtime && CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/main.go

EXPOSE 8080

CMD ["/app/app"]
