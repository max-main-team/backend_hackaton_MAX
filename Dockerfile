FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o app ./cmd/uni_bot

FROM golang:1.25-alpine

WORKDIR /root/

COPY --from=builder /app/app /root/app
COPY --from=builder /app/cfg/config.toml ./cfg/config.toml
COPY --from=builder docs/swagger.json  ./docs/swagger.json 
COPY --from=builder docs/swagger.yaml  ./docs/swagger.yaml 

EXPOSE 8080

CMD ["./app"]