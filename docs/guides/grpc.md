# gRPC Guide

This project runs a gRPC server alongside the HTTP server. The two share the same domain and use-case layer — you wire the same use-case instance into both HTTP handlers and gRPC handlers.

---

## Proto Repository

Proto files live in a **separate repository**:
[github.com/arisatriop/goilerplate-proto](https://github.com/arisatriop/goilerplate-proto)

```
goilerplate-proto/
  buf.yaml          # buf module config, lint rules, BSR dependencies
  buf.gen.yaml      # code generation config (plugins + output paths)
  buf.lock          # pinned BSR dependency versions
  bar/v1/
    bar.proto
    bar.pb.go           # generated — do not edit
    bar_grpc.pb.go      # generated — do not edit
  foo/v1/
  hello/v1/
```

This repo is the single source of truth for all service contracts. Both the server (goilerplate) and any client service import from here.

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

## Adding a New gRPC Service

Adding a new service (e.g. `baz`) involves two repos.

### Step 1 — Add proto to goilerplate-proto

In the `goilerplate-proto` repo, create `baz/v1/baz.proto`:

```proto
syntax = "proto3";

package baz.v1;

import "google/api/field_behavior.proto";
import "google/protobuf/empty.proto";

option go_package = "github.com/arisatriop/goilerplate-proto/baz/v1";

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

Then generate, commit, and tag a new version:

```bash
# in goilerplate-proto/
buf generate
git add -A && git commit -m "feat: add BazService proto"
git tag v0.2.0 && git push origin main --tags
```

### Step 2 — Update goilerplate to use the new version

```bash
# in goilerplate/
go get github.com/arisatriop/goilerplate-proto@v0.2.0
go mod tidy
```

### Step 3 — Write the handler

Create `internal/delivery/grpc/handler/baz.go`:

```go
package grpchandler

import (
    "context"

    bazdomain "goilerplate/internal/domain/baz"
    "goilerplate/pkg/grpcresponse"
    pb "github.com/arisatriop/goilerplate-proto/baz/v1"

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

func toProtoBaz(e *bazdomain.Baz) *pb.Baz {
    return &pb.Baz{Id: e.ID, Name: e.Name}
}
```

### Step 4 — Register in the service registry

In `internal/delivery/grpc/server.go`:

```go
import bazpb "github.com/arisatriop/goilerplate-proto/baz/v1"

type ServiceRegistry struct {
    // ...
    Baz *grpchandler.Baz
}

func (r *ServiceRegistry) Register(s *grpc.Server) {
    // ...
    bazpb.RegisterBazServiceServer(s, r.Baz)
}
```

### Step 5 — Wire it

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

For simple input validation inside the handler:

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

## Local Testing with grpcurl

Server reflection is **disabled** (production mode). You must provide the proto files explicitly when using grpcurl.

### Install grpcurl

```bash
brew install grpcurl
```

### Setup googleapis path (one-time)

buf caches googleapis locally after running `buf generate`. Find the path:

```bash
find ~/.cache/buf -name "field_behavior.proto" 2>/dev/null | head -1
# Example: ~/.cache/buf/v3/modules/.../files/google/api/field_behavior.proto
# GOOGLEAPIS = everything up to /files
```

Export for convenience:

```bash
export GOILERPLATE_PROTO=~/Documents/work/others/goilerplate-proto
export GOOGLEAPIS=~/.cache/buf/v3/modules/b5/buf.build/googleapis/googleapis/<commit>/files
```

### Call methods

```bash
# ListBars
grpcurl -plaintext \
  -import-path $GOILERPLATE_PROTO \
  -import-path $GOOGLEAPIS \
  -proto bar/v1/bar.proto \
  -d '{"page":1,"limit":10}' \
  127.0.0.1:50051 \
  bar.v1.BarService/ListBars

# CreateBar
grpcurl -plaintext \
  -import-path $GOILERPLATE_PROTO \
  -import-path $GOOGLEAPIS \
  -proto bar/v1/bar.proto \
  -d '{"code":"EXP001","bar":"My Bar"}' \
  127.0.0.1:50051 \
  bar.v1.BarService/CreateBar

# GetBar
grpcurl -plaintext \
  -import-path $GOILERPLATE_PROTO \
  -import-path $GOOGLEAPIS \
  -proto bar/v1/bar.proto \
  -d '{"id":"<uuid>"}' \
  127.0.0.1:50051 \
  bar.v1.BarService/GetBar
```

> **Note:** Use `127.0.0.1:50051` instead of `localhost:50051` to avoid IPv6 resolution issues on macOS.

### Pass metadata (service-to-service)

```bash
grpcurl -plaintext \
  -import-path $GOILERPLATE_PROTO \
  -import-path $GOOGLEAPIS \
  -proto bar/v1/bar.proto \
  -H "x-service-name: my-service" \
  -H "x-request-id: abc-123" \
  -d '{"code":"EXP001","bar":"My Bar"}' \
  127.0.0.1:50051 \
  bar.v1.BarService/CreateBar
```

---

## Service-to-Service Calls

Client services import the same proto module and connect directly:

```go
import (
    barpb "github.com/arisatriop/goilerplate-proto/bar/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
)

conn, err := grpc.NewClient("127.0.0.1:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
client := barpb.NewBarServiceClient(conn)

ctx = metadata.AppendToOutgoingContext(ctx,
    "x-service-name", "my-service",
    "x-request-id",   requestID,
)

resp, err := client.CreateBar(ctx, &barpb.CreateBarRequest{
    Code: "EXP001",
    Bar:  "My Bar",
})
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
