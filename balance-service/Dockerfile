FROM golang:1.23 AS builder

WORKDIR /app/

RUN apt-get update && apt-get install -y librdkafka-dev

COPY go.mod go.sum ./ 
RUN go mod download

COPY . .

WORKDIR /app

RUN go build -o /app/bin/balance-service .

FROM golang:1.23

WORKDIR /root/

COPY --from=builder /app/bin/balance-service .

EXPOSE 3003 

CMD ["./balance-service"]
