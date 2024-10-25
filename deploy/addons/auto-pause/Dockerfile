FROM golang:1.23.2 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN GOOS=linux CGO_ENABLED=0 go build -a --ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo -o auto-pause-hook cmd/auto-pause/auto-pause-hook/main.go cmd/auto-pause/auto-pause-hook/config.go cmd/auto-pause/auto-pause-hook/certs.go

FROM scratch
COPY --from=builder /app/auto-pause-hook /auto-pause-hook
