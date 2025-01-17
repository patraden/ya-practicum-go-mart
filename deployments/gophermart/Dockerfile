# Use Go 1.22.8 as the base image
FROM golang:1.22.8 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o gophermart ./cmd/gophermart

# Create a lightweight final image
FROM golang:1.22.8

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/gophermart /usr/local/bin/gophermart

CMD /usr/local/bin/gophermart -a "${GOPHERMART_ADDRESS}" -d "${DSN}"