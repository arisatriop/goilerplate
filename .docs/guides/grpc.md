# gRPC Guide

This project runs a gRPC server alongside the HTTP server. The two share the same domain and use-case layer — you wire the same use-case instance into both HTTP handlers and gRPC handlers.

---

## Configuration

The `grpc` block is already present in `config/config.example.yaml`. Copy it to your local `config/config.yaml` if it's missing:

```yaml
grpc:
  enabled: true
  port: 50051
```

The gRPC server only starts when `enabled: true`. The HTTP server always starts regardless.

---

## Proto Structure

Proto files live in `proto/` and are organized by service and version:

```
proto/
  buf.yaml          # buf module config, lint rules, BSR dependencies
  buf.gen.yaml      # code generation config (plugins + output paths)
  buf.lock          # pinned BSR dependency versions
  bar/
    v1/
      bar.proto
      bar.pb.go           # generated — do not edit
      bar_grpc.pb.go      # generated — do not edit
  foo/
    v1/
      ...
  hello/
    v1/
      ...
```

Generated Go files (`*.pb.go`, `*_grpc.pb.go`) live next to their `.proto` source because `buf.gen.yaml` uses `paths=source_relative`.

---

## Generating Code

```bash
make proto-gen    # runs: cd proto && buf generate
make proto-lint   # runs: cd proto && buf lint
```

Requires `buf` CLI — install with `brew install bufbuild/buf/buf`.

Dependencies (e.g. `google/api/field_behavior.proto`) are resolved from the Buf Schema Registry (BSR) — no local copies needed.

---

## Adding a New gRPC Service

Follow these steps when adding a new service (e.g. `baz`):

### 1. Write the proto

Create `proto/baz/v1/baz.proto`:

```proto
syntax = "proto3";

package baz.v1;

import "google/api/field_behavior.proto";
import "google/protobuf/empty.proto";

option go_package = "goilerplate/proto/baz/v1";

service BazService {
  rpc CreateBaz (CreateBazRequest) returns (Baz);
  rpc GetBaz    (GetBazRequest)    returns (Baz);
  rpc ListBazs  (ListBazsRequest)  returns (ListBazsResponse);
  rpc UpdateBaz (UpdateBazRequest) returns (Baz);
  rpc DeleteBaz (DeleteBazRequest) returns (google.protobuf.Empty);
}

message Baz {
  string id   = 1 [(google.api.field_behavior) = OUTPUT_ONLY];
  string name = 2 [(google.api.field_behavior) = OUTPUT_ONLY];
}

// ... request/response messages
```

Then generate:

```bash
make proto-gen
```

### 2. Write the handler

Create `internal/delivery/grpc/handler/baz.go`:

```go
package grpchandler

import (
    "context"

    bazdomain "goilerplate/internal/domain/baz"
    "goilerplate/pkg/grpcresponse"
    pb "goilerplate/proto/baz/v1"

    "google.golang.org/protobuf/types/known/emptypb"
)

type Baz struct {
    pb.UnimplementedBazServiceServer
    uc bazdomain.Usecase
}

func NewBaz(uc bazdomain.Usecase) *Baz {
    return &Baz{uc: uc}
}

func (b *Baz) CreateBaz(ctx context.Context, req *pb.CreateBazRequest) (*pb.Baz, error) {
    entity := &bazdomain.Baz{Name: req.Name}
    created, err := b.uc.Create(ctx, entity)
    if err != nil {
        return nil, grpcresponse.HandleError(ctx, err)
    }
    return toProtoBaz(created), nil
}

// ... other methods

func toProtoBaz(e *bazdomain.Baz) *pb.Baz {
    return &pb.Baz{Id: e.ID, Name: e.Name}
}
```

### 3. Register in the service registry

In `internal/delivery/grpc/server.go`, add the new service:

```go
import (
    bazpb "goilerplate/proto/baz/v1"
)

type ServiceRegistry struct {
    // ...
    Baz *grpchandler.Baz
}

func (r *ServiceRegistry) Register(s *grpc.Server) {
    // ...
    bazpb.RegisterBazServiceServer(s, r.Baz)
}
```

### 4. Wire it

In `internal/wire/handler_grpc.go`:

```go
baz := grpchandler.NewBaz(useCases.BazUC)
registry := grpcdelivery.NewServiceRegistry(hello, foo, bar, baz)
```

---

## Error Handling

Use `pkg/grpcresponse` to translate domain errors to gRPC status codes:

```go
if err != nil {
    return nil, grpcresponse.HandleError(ctx, err)
}
```

Domain errors (e.g. `ErrNotFound`, `ErrAlreadyExists`) are mapped to the appropriate `codes.NotFound`, `codes.AlreadyExists`, etc.

For simple input validation errors inside the handler:

```go
import "google.golang.org/grpc/codes"
import "google.golang.org/grpc/status"

if req.Id == "" {
    return nil, status.Error(codes.InvalidArgument, "id is required")
}
```

---

## Middleware

The gRPC server uses two interceptors (configured in `internal/bootstrap/grpc.go`):

| Interceptor | Purpose |
|---|---|
| `RequestLogger` | Logs method, peer, request, response, latency |
| `Recovery` | Catches panics and returns `codes.Internal` |

`RequestLogger` also injects two values into context:

- `constants.ContextKeyRequestID` — from `x-request-id` metadata, or a new UUID
- `constants.ContextKeyUserID` — from `x-service-name` metadata, or `"system"`

This means any use-case that reads the caller identity from context will work for both HTTP and gRPC calls without changes.

---

## Reflection & Local Testing

Server reflection is enabled in all non-production environments (`app.env != production`). Reflection lets clients discover services without a `.proto` file.

### Install grpcurl

```bash
brew install grpcurl
```

### List available services

```bash
grpcurl -plaintext localhost:50051 list
```

### Call a method

```bash
# CreateBar
grpcurl -plaintext \
  -d '{"code":"B001","bar":"My Bar"}' \
  localhost:50051 \
  bar.v1.BarService/CreateBar

# GetBar
grpcurl -plaintext \
  -d '{"id":"<uuid>"}' \
  localhost:50051 \
  bar.v1.BarService/GetBar

# ListBars
grpcurl -plaintext \
  -d '{"page":1,"limit":10}' \
  localhost:50051 \
  bar.v1.BarService/ListBars

# UpdateBar
grpcurl -plaintext \
  -d '{"id":"<uuid>","code":"B002","bar":"Updated Bar"}' \
  localhost:50051 \
  bar.v1.BarService/UpdateBar

# DeleteBar
grpcurl -plaintext \
  -d '{"id":"<uuid>"}' \
  localhost:50051 \
  bar.v1.BarService/DeleteBar
```

### Pass metadata (service-to-service)

```bash
grpcurl -plaintext \
  -H "x-service-name: my-service" \
  -H "x-request-id: abc-123" \
  -d '{"code":"B001","bar":"My Bar"}' \
  localhost:50051 \
  bar.v1.BarService/CreateBar
```

---

## Service-to-Service Calls

When another service calls this gRPC server, it should pass its service name via metadata so the request logger can identify the caller:

```go
import "google.golang.org/grpc/metadata"

md := metadata.Pairs(
    "x-service-name", "payment-service",
    "x-request-id",   requestID,
)
ctx = metadata.NewOutgoingContext(ctx, md)

resp, err := client.CreateBar(ctx, req)
```

If `x-service-name` is absent, the caller is recorded as `"system"`.

---

## Proto Design Conventions

This project follows [Google AIP](https://google.aip.dev/) standards:

| Operation | AIP | Return type |
|---|---|---|
| Create | [AIP-133](https://google.aip.dev/133) | Resource directly |
| Get | [AIP-131](https://google.aip.dev/131) | Resource directly |
| List | [AIP-132](https://google.aip.dev/132) | `List<Resource>Response` |
| Update | [AIP-134](https://google.aip.dev/134) | Resource directly |
| Delete | [AIP-135](https://google.aip.dev/135) | `google.protobuf.Empty` |

Use `google.api.field_behavior` annotations as documentation hints:

```proto
string id   = 1 [(google.api.field_behavior) = OUTPUT_ONLY];
string code = 2 [(google.api.field_behavior) = REQUIRED];
int32  page = 3 [(google.api.field_behavior) = OPTIONAL];
```

These are metadata only — they do not enforce validation at runtime. Validation is the handler's responsibility.
