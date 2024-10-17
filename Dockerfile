FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app ./cmd/server/main.go

# make sure .env file exists
RUN touch .env

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/app ./

COPY --from=builder /app/.env ./

EXPOSE 3000
CMD ["./app"]