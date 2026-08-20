FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/certpilot-api ./cmd/api

FROM alpine:3.20
RUN addgroup -S certpilot && adduser -S certpilot -G certpilot
WORKDIR /app
COPY --from=build /out/certpilot-api /app/certpilot-api
COPY configs/config.example.yaml /app/config.yaml
USER certpilot
EXPOSE 8080
ENTRYPOINT ["/app/certpilot-api", "--config", "/app/config.yaml"]
