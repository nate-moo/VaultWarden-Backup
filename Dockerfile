FROM golang:1.24 AS builder
LABEL authors="nathan"

WORKDIR /app
COPY main.go go.mod go.sum /app/

RUN go build .

FROM busybox:latest AS runner

WORKDIR /app

COPY --from=builder /app/VaultwardenBackup /app/

ENTRYPOINT ["/app/VaultwardenBackup"]
