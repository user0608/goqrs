FROM golang:1.20.4-alpine AS builder
RUN apk add git
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o build/main main.go

FROM alpine
WORKDIR /app
COPY  --from=builder /app/build/main .
CMD ["./main"]