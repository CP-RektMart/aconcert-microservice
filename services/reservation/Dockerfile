FROM golang:1.25.0-alpine3.21

WORKDIR /app

COPY go.mod go.sum ./
COPY services/reservation/go.mod services/reservation/go.sum ./services/reservation/
COPY pkg ./pkg/

RUN cd services/reservation && go mod download

COPY services/reservation ./services/reservation/

RUN cd services/reservation && CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/main.go

EXPOSE 8080

CMD ["/app/app"]
