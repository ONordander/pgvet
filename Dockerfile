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
LABEL com.github.actions.name="pgvet" \
  com.github.actions.description="Lint PostgreSQL migration scripts" \
  maintainer="@ONordander" \
  org.opencontainers.image.url="https://github.com/ONordander/pgvet" \
  org.opencontainers.image.source="https://github.com/ONordander/pgvet"

WORKDIR /
COPY --from=builder /build/default-config.yaml .
COPY --from=builder /build/pgvet .
ENTRYPOINT ["/pgvet", "lint", "--exit-status-on-violation"]
