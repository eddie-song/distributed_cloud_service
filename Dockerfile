# syntax=docker/dockerfile:1
FROM golang:1.22 as builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/node ./cmd/node
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/cloudctl ./cmd/cloudctl

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/node /app/node
COPY --from=builder /out/cloudctl /app/cloudctl
COPY web /app/web
COPY configs /app/configs
ENV PATH=/app:$PATH
ENTRYPOINT ["/app/node"]
