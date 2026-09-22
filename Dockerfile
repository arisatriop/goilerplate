# syntax=docker/dockerfile:1

# ── Build stage ──────────────────────────────────────────────────────────────
# Pinned to the toolchain in go.mod. A floating tag means two builds of one
# commit can be compiled by different compilers, which is the opposite of what a
# deployable artefact is for.
FROM golang:1.26.8-alpine AS builder

WORKDIR /app

# Dependencies first, so a source-only change does not re-download the module
# cache. --mount keeps the caches out of the image layers entirely.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

# Stamped at build time and read back by config.App.Version, so a running
# container can say exactly which commit it is.
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# CGO_ENABLED=0 produces a static binary, which is what lets the runtime stage
# be distroless/static rather than a full distribution.
#
# -trimpath strips the build machine's absolute paths out of the binary; without
# it the same source compiled in two checkouts produces two different binaries.
# -s -w drop the symbol table and DWARF data — roughly a third of the size, at
# the cost of symbol names in a panic trace, which a stack trace from a stripped
# Go binary still survives.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w \
        -X 'main.version=${VERSION}' \
        -X 'main.commit=${COMMIT}' \
        -X 'main.buildDate=${BUILD_DATE}'" \
      -o /out/goilerplate ./cmd/server

# ── Runtime stage ────────────────────────────────────────────────────────────
# distroless/static rather than alpine: no shell, no package manager, no libc,
# nothing to update and nothing for an attacker who reaches RCE to pivot with.
# It is pinned by digest — `latest` on a base image makes the build
# unreproducible and silently changes what ships.
#
# :nonroot runs as uid 65532 and needs no adduser step.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab

WORKDIR /app

COPY --from=builder /out/goilerplate /app/goilerplate

USER nonroot:nonroot
EXPOSE 3000

ENTRYPOINT ["/app/goilerplate"]
