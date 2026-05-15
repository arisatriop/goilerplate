---
name: grpc-guide
description: How to add or modify gRPC services in this project. Use when touching proto contracts, gRPC handlers in internal/delivery/grpc/, the service registry, the goilerplate-proto module, grpc config, or when testing gRPC locally with grpcurl.
---

# gRPC Guide

The gRPC server runs alongside the HTTP server and shares the same domain and
use-case layer — the same use-case instance is wired into both HTTP and gRPC
handlers. Full reference with proto/grpcurl examples: `docs/guides/grpc.md`.

## Proto contract lives in a separate repo

Proto files are in [github.com/arisatriop/goilerplate-proto](https://github.com/arisatriop/goilerplate-proto),
the single source of truth for all service contracts. Generated `*.pb.go` files
are never edited by hand. Server reflection is **disabled** — clients must import
the proto module.

## Adding a new gRPC service (e.g. `baz`)

1. **Proto** — in `goilerplate-proto`, add `baz/v1/baz.proto`, run `buf generate`,
   commit, then tag a new version (`git tag v0.2.0 && git push origin main --tags`).
2. **Pull the version** — in goilerplate: `go get github.com/arisatriop/goilerplate-proto@v0.2.0 && go mod tidy`.
3. **Handler** — create `internal/delivery/grpc/handler/baz.go`: a struct embedding
   `pb.UnimplementedBazServiceServer` plus the domain `Usecase`, a `NewBaz` constructor,
   and one method per RPC. Map domain entities to/from proto messages with helper funcs.
4. **Register** — in `internal/delivery/grpc/server.go`: add the handler to
   `ServiceRegistry` and call `bazpb.RegisterBazServiceServer(s, r.Baz)` in `Register`.
5. **Wire** — in `internal/wire/handler_grpc.go`: construct the handler and pass it
   to `NewServiceRegistry`.

## Conventions

- Translate domain errors with `grpcresponse.HandleError(ctx, err)`; use
  `status.Error(codes.InvalidArgument, ...)` for simple in-handler validation.
- Two interceptors run (`internal/bootstrap/grpc.go`): `RequestLogger` (logs +
  injects request ID / user ID into context) and `Recovery` (panic → `codes.Internal`).
  Because identity is injected into context, use-cases work unchanged for HTTP and gRPC.
- Follow [Google AIP](https://google.aip.dev/): Create/Get/Update return the resource
  directly, List returns `List<Resource>Response`, Delete returns `google.protobuf.Empty`.
- `google.api.field_behavior` annotations are documentation hints only — runtime
  validation is still the handler's responsibility.

## Local testing

`grpcurl` requires explicit proto paths since reflection is off. See
`docs/guides/grpc.md` for the full `grpcurl` invocations and the googleapis path
setup. Use `127.0.0.1:50051`, not `localhost`, to avoid IPv6 issues on macOS.
