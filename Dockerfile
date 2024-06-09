FROM golang:1.22.3-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN GOOS=linux go build -ldflags="-s" -o emailer

FROM golang:1.22.3-alpine
COPY --from=builder /app/emailer /emailer
EXPOSE 8080
ENTRYPOINT ["/emailer"]