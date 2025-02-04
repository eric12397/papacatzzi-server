FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o main .

FROM ubuntu:22.04
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
