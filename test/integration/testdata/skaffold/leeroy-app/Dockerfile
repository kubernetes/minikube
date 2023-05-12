ARG BASE
FROM golang:1.18 as builder
WORKDIR /code
COPY app.go .
COPY go.mod .
# `skaffold debug` sets SKAFFOLD_GO_GCFLAGS to disable compiler optimizations
ARG SKAFFOLD_GO_GCFLAGS
RUN go build -gcflags="${SKAFFOLD_GO_GCFLAGS}" -trimpath -o /app .

FROM $BASE
COPY --from=builder /app .
