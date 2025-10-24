# Build
FROM golang:1.24.3 AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /app/users-api ./main.go

# Run
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/users-api /usr/local/bin/users-api
ENV DB_CONNECTION_URL=""
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/users-api"]
