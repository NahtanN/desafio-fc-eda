FROM golang:1.23 AS builder

WORKDIR /app/

RUN apt-get update && apt-get install -y librdkafka-dev

COPY go.mod go.sum ./ 
RUN go mod download

COPY . .

WORKDIR /app/cmd/walletcore

RUN go build -o /app/bin/walletcore .

FROM golang:1.23

WORKDIR /root/

COPY --from=builder /app/bin/walletcore .

EXPOSE 8080

CMD ["./walletcore"]
