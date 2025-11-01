FROM golang:1.25.0-alpine3.21

WORKDIR /app

COPY go.mod go.sum ./
COPY services/location/go.mod services/location/go.sum ./services/location/
COPY pkg ./pkg/

RUN cd services/location && go mod download

COPY services/location ./services/location/

RUN cd services/location && CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/main.go

EXPOSE 8080

CMD ["/app/app"]
