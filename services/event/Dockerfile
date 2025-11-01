FROM golang:1.25.0-alpine3.21

WORKDIR /app

COPY go.mod go.sum ./
COPY services/event/go.mod services/event/go.sum ./services/event/
COPY pkg ./pkg/

RUN cd services/event && go mod download

COPY services/event ./services/event/

RUN cd services/event && CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/main.go

EXPOSE 8080

CMD ["/app/app"]
