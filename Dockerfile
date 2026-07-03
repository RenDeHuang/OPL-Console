FROM node:22-bookworm AS web
WORKDIR /src
COPY apps/web/package*.json apps/web/
RUN npm --prefix apps/web ci
COPY apps/web apps/web
RUN npm --prefix apps/web run build

FROM golang:1.23-bookworm AS api
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd cmd
COPY internal internal
COPY migrations migrations
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/console-api ./cmd/console-api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=api /out/console-api /app/console-api
COPY --from=web /src/apps/web/dist /app/apps/web/dist
COPY migrations /app/migrations
EXPOSE 8787
USER nonroot:nonroot
ENTRYPOINT ["/app/console-api"]
