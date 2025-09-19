FROM golang:1.25-alpine AS base

FROM base AS dev

# Needed for go install and private repos if any
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Install Air (official module path)
RUN go install github.com/air-verse/air@latest

# Cache deps early
COPY go.mod go.sum ./
RUN go mod download

# Bring in the rest of the code
COPY . .

# Default port — adjust to your chi server’s port
EXPOSE 8080

# Use a non-root user for local dev safety (optional)
RUN adduser -D -u 10001 appuser
USER appuser

# Air will read .air.toml by default; explicitly pass it for clarity
CMD ["air", "-c", ".air.toml"]



# FROM golang:1.2 AS build
# WORKDIR /src
# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server


# FROM gcr.io/distroless/static:nonroot
# WORKDIR /
# COPY --from=build /out/server /server
# EXPOSE 8080
# USER nonroot:nonroot
# ENTRYPOINT ["/server"]
