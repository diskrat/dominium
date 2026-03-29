# Compilation
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/sim/main.go

# Execution
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web
EXPOSE 8080 8081 8082
ENTRYPOINT ["./main"]