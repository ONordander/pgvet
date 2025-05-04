FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY NOTICE.txt .
COPY VERSION.txt .
COPY go.mod .
COPY go.sum .
COPY *.go .
COPY ./rules ./rules

# Create a default config that enables everything
RUN echo "rules:" >> default-config.yaml

RUN CGO_ENABLED=0 go build .

FROM scratch
LABEL com.github.actions.name="pgcheck" \
  com.github.actions.description="Lint PostgreSQL migration scripts" \
  maintainer="@ONordander" \
  org.opencontainers.image.url="https://github.com/ONordander/pgcheck" \
  org.opencontainers.image.source="https://github.com/ONordander/pgcheck"

WORKDIR /
COPY --from=builder /build/default-config.yaml .
COPY --from=builder /build/pgcheck .
ENTRYPOINT ["/pgcheck", "lint", "--exit-status-on-violation"]
