FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

RUN go build -o app ./cmd/uni_bot

FROM golang:1.25-alpine

WORKDIR /root/

COPY --from=builder /app/app /root/app

EXPOSE 8080

CMD ["./app"]