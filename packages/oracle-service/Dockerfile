FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /oracle-service ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

COPY --from=builder /oracle-service /usr/local/bin/oracle-service

EXPOSE 50053

CMD ["oracle-service"]
