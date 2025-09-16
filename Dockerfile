FROM golang:1.25.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v ./...

FROM golang:1.25.1-alpine AS test

WORKDIR /app

RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "test", "-v", "./..."]

FROM alpine:latest AS runtime

RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

COPY --from=builder /app .

CMD ["./ddao"]
