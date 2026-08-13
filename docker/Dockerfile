FROM golang:1.25-alpine AS build

RUN apk --no-cache add git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o /app/server ./cmd

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=build /app/server /app/server

RUN mkdir -p /app/config/logs && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8888

ENTRYPOINT ["/app/server"]