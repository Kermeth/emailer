FROM golang:1.23.12-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN GOOS=linux go build -ldflags="-s" -o emailer

FROM golang:1.23.12-alpine
COPY --from=builder /app/emailer /emailer
EXPOSE 8080
ENTRYPOINT ["/emailer"]