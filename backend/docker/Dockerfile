FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/bin/backend ./cmd/backend
RUN CGO_ENABLED=0 GOBIN=/app/bin go install -tags mysql github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.0

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/bin/backend ./backend
COPY --from=build /app/bin/migrate /usr/local/bin/migrate
COPY migrations ./migrations
COPY scripts/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./backend"]
