# --- build stage ---
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@latest

# --- runtime stage ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/api /app/main
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY db /app/db
EXPOSE 8080
ENTRYPOINT ["/bin/sh", "-c", "goose -dir /app/db/migrations postgres \"$DATABASE_URL\" up && exec /app/main"]
