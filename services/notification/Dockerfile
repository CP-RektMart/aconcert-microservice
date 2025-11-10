FROM golang:1.25.0-alpine3.21

WORKDIR /app

COPY go.mod go.sum ./
COPY services/notification/go.mod services/notification/go.sum ./services/notification/
COPY pkg ./pkg/

RUN cd services/notification && go mod download

COPY services/notification ./services/notification/

RUN cd services/notification && CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/main.go

CMD ["/app/app"]
