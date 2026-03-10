# CRUD Operations Guide

**Step-by-step guide to create new CRUD operations. Use this guide with AI Agent to build new features.**

> **Note:** Replace `Foo` with your entity name (e.g. `Product`, `Order`, `Customer`). Use PascalCase for struct names and lowercase for folder/file names.

---

## 1. Define the Domain

Create a new folder in `internal/domain` with your domain name. For example `foo`, create the following files:

1. **`error.go`** - Define error types if needed (Optional)
2. **`filter.go`** - Define filter struct if needed (Optional)
3. **`message.go`** - Define message constants if needed (Optional)
4. **`entity.go`** - Define entity struct
5. **`repository.go`** - Define repository interface with CRUD methods:
   - `WithTx(ctx context.Context) Repository`
   - `CreateFoo(ctx context.Context, entities *Foo) (*Foo, error)`
   - `UpdateFoo(ctx context.Context, entities *Foo) error`
   - `DeleteFoo(ctx context.Context, entities *Foo) error`
   - `BulkCreate(ctx context.Context, entities []*Foo) error`
   - `CountFoo(ctx context.Context, filter *Filter) (int64, error)`
   - `GetFooList(ctx context.Context, filter *Filter) ([]*Foo, error)`
   - `GetFooByID(ctx context.Context, id string) (*Foo, error)`
6. **`usecase.go`** - Define usecase interface and its implementation:
   - `Create(ctx context.Context, entity *Foo) error`
   - `Update(ctx context.Context, entity *Foo) error`
   - `Delete(ctx context.Context, entity *Foo) error`
   - `BulkCreate(ctx context.Context, entities []*Foo) error`
   - `Count(ctx context.Context, filter *Filter) (int64, error)`
   - `GetList(ctx context.Context, filter *Filter) ([]*Foo, error)`
   - `GetByID(ctx context.Context, id string) (*Foo, error)`

---

## 2. Define the Model

Create file `foo.go` in `internal/infrastructure/model` with the following struct:

```go
type Foo struct {
    ID        string     `gorm:"primaryKey;default:gen_random_uuid()"`
    Code      string     `gorm:"column:code"`
    Name      string     `gorm:"column:name"`
    IsActive  bool       `gorm:"column:is_active"`
    CreatedBy string     `gorm:"column:created_by"`
    UpdatedBy string     `gorm:"column:updated_by"`
    DeletedBy *string    `gorm:"column:deleted_by"`
    CreatedAt time.Time  `gorm:"column:created_at"`
    UpdatedAt time.Time  `gorm:"column:updated_at"`
    DeletedAt *time.Time `gorm:"column:deleted_at"`
}
```

---

## 3. Define Repository Implementation

Create file `foo.go` in `internal/infrastructure/repository` to implement the repository interface:

- `fooRepo` - Repository struct
- `NewFooRepo(db *gorm.DB) foo.Repository` - Constructor
- `WithTx(ctx context.Context) foo.Repository` - Transaction method
- `CreateFoo(ctx context.Context, entity *Foo) (*foo.Foo, error)`
- `UpdateFoo(ctx context.Context, entity *Foo) error`
- `DeleteFoo(ctx context.Context, entity *Foo) error`
- `BulkCreate(ctx context.Context, entities []*Foo) error`
- `CountFoo(ctx context.Context, filter *Filter) (int64, error)`
- `GetFooList(ctx context.Context, filter *Filter) ([]*foo.Foo, error)`
- `GetFooByID(ctx context.Context, id string) (*foo.Foo, error)`

---

## 4. Define Request & Response DTOs

### 4.1 Request DTOs

Create file `foo.go` in `internal/delivery/http/dto/request`:

- `FooCreateRequest`
- `FooUpdateRequest`
- `FooListRequest`

### 4.2 Response DTOs

Create file `foo.go` in `internal/delivery/http/dto/response`:

- `FooResponse`

### 4.3 Request Parser

Create file `foo.go` in `internal/delivery/http/request`:

- `func ToFooFilter(req *dtorequest.FooListRequest, ctx *fiber.Ctx) *foo.Filter`

### 4.4 Presenter

Create file `foo.go` in `internal/delivery/http/presenter`:

- `func ToFooResponse(entity *foo.Foo) *dtoresponse.FooResponse`
- `func ToFooListResponse(entities []*foo.Foo) []*dtoresponse.FooResponse`

---

## 5. Define the Handler

Create file `foo.go` in `internal/delivery/http/handler`:

- `Foo` - Handler struct
- `NewFoo(validator *validator.Validate, usecase foo.Usecase) *Foo` - Constructor
- `func(h *Foo) Create(ctx *fiber.Ctx) error`
- `func(h *Foo) Update(ctx *fiber.Ctx) error`
- `func(h *Foo) Delete(ctx *fiber.Ctx) error`
- `func(h *Foo) GetList(ctx *fiber.Ctx) error`
- `func(h *Foo) GetByID(ctx *fiber.Ctx) error`

---

## 6. Add Wire Binding

Setup Dependency Injection in `internal/wire/`:

1. **Bind repository** in `internal/wire/repository.go`
2. **Bind usecase** in `internal/wire/usecase.go`
3. **Bind handler** in `internal/wire/handler.go`

---

## 7. Add Permissions

Add permission constants in `pkg/constants/permission.go`:

```go
const (
    PermissionFooCreate = "foo.create"
    PermissionFooRead   = "foo.read"
    PermissionFooUpdate = "foo.update"
    PermissionFooDelete = "foo.delete"
)
```

---

## 8. Add API Route

Register API route based on scope:

1. **Public API** (end-user) → `internal/delivery/http/router/public.go`
2. **Partner API** (third-party) → `internal/delivery/http/router/partner.go`
3. **Internal API** (service-to-service) → `internal/delivery/http/router/internal.go`

---

## 9. Important Notes ⚠️

**If API requires data from more than 1 domain**, implement application layer for orchestrating multiple domains.

Example: If creating an order involving User + Product:
- Domain layer: `user`, `product`, `order` (separate)
- Application layer: `CreateOrderService` (orchestrates all domains)
- Handler: Call application service, not usecase directly

---

## Summary Checklist

- [ ] Domain: entity.go, repository.go, usecase.go
- [ ] Model: infrastructure/model/foo.go
- [ ] Repository: infrastructure/repository/foo.go
- [ ] DTOs: delivery/http/dto/{request,response}/foo.go
- [ ] Converters: delivery/http/{request,presenter}/foo.go
- [ ] Handler: delivery/http/handler/foo.go
- [ ] Wire bindings: wire/{repository,usecase,handler}.go
- [ ] Permissions: pkg/constants/permission.go
- [ ] Routes: router/{public,partner,internal}.go
