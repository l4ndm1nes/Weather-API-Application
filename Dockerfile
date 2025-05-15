FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o weather-api-application ./cmd/httpserver

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/weather-api-application .
EXPOSE 8080
CMD ["./weather-api-application"]
