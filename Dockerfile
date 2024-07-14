# Stage 1: Build the Go binary
FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN go build -o main main.go

# Stage 2: Run the Go binary
FROM alpine

WORKDIR /app

COPY --from=builder /app/main .
COPY .env .

# Ensure the binary is executable
RUN chmod +x main

CMD ["./main"]
