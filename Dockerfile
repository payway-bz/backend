FROM golang:1.25-alpine AS base

WORKDIR /app

# Default to static binaries; flip to 1 if you need CGO
ENV CGO_ENABLED=0

FROM base AS make-lockfile

COPY go.mod ./

# Bring in the rest of the code
COPY cmd cmd
COPY internal internal

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod tidy && \
    go mod download && \
    go mod verify


FROM base AS build

# Freeze module changes everywhere after this point (for build/prod flows)
ENV GOFLAGS="-mod=readonly"

# Cache deps early
COPY go.mod go.sum ./
RUN go mod download

# Bring in the rest of the code
COPY cmd cmd
COPY internal internal


# Build the static binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags='-s -w -extldflags "-static"' -o /out/app ./cmd/server

# Default port — adjust to your chi server’s port
EXPOSE 8080

FROM build AS dev

# In dev, allow resolving and updating go.sum as code changes
ENV GOFLAGS=""

# Install Air (official module path)
RUN go install github.com/air-verse/air@latest

# Air will read .air.toml by default; explicitly pass it for clarity
CMD ["air", "-c", ".air.toml"]


# Distroless keeps it minimal and non-root by default
FROM gcr.io/distroless/static:nonroot AS prod
WORKDIR /app

# Copy the statically linked binary from the build stage
COPY --from=build /out/app /app/app

# Match the port your server listens on
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
