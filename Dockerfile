# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/seeder ./cmd/seeder/main.go

# Run stage
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl

COPY --from=builder /app/server /app/server
COPY --from=builder /app/seeder /app/seeder
COPY toll_plaza_india.csv /app/toll_plaza_india.csv
COPY migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/server"]
