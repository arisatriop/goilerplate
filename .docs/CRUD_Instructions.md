# CRUD Instructions 
**Below are step by step to create new CRUD operations. You can use this instruction to create new CRUD operations for AI Agent.**

> **Note:** Replace `Foo` with your actual entity name (e.g. `Product`, `Order`, `Customer`). Use PascalCase for struct names and lowercase for folder/file names.

## 1. Define the domain
Create a new folder in the `internal/domain` directory with the name of your domain. Inside the folder e.g. `foo`, create the following files:
1. `error.go`: Define the error types if needed (Optional).
2. `filter.go`: Define the filter struct if needed (Optional).
3. `message.go`: Define the message constants if needed (Optional).
4. `entity.go`: Define the entity struct.
5. `repository.go`: Define the repository interface with the basic CRUD methods:
    - `WithTx(ctx context.Context) Repository`
    - `CreateFoo(ctx context.Context, entities *Foo) (*Foo, error)`
    - `UpdateFoo(ctx context.Context, entities *Foo) error`
    - `DeleteFoo(ctx context.Context, entities *Foo) error`
    - `BulkCreate(ctx context.Context, entities []*Foo) error`
    - `CountFoo(ctx context.Context, filter *Filter) (int64, error)`
    - `GetFooList(ctx context.Context, filter *Filter) ([]*Foo, error)`
    - `GetFooByID(ctx context.Context, id string) (*Foo, error)`
6. `usecase.go`: Define the usecase interface and its implementation with the basic CRUD methods:
    - `Create(ctx context.Context, entity *Foo) error`
    - `Update(ctx context.Context, entity *Foo) error`
    - `Delete(ctx context.Context, entity *Foo) error`
    - `BulkCreate(ctx context.Context, entities []*Foo) error`
    - `Count(ctx context.Context, filter *Filter) (int64, error)`
    - `GetList(ctx context.Context, filter *Filter) ([]*Foo, error)`
    - `GetByID(ctx context.Context, id string) (*Foo, error)`

## 2. Define the model 
Create a new file e.g. `foo.go` in the `internal/infrastructure/model` directory, then define the struct with basic field as follow:
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

## 3. Define the repository implementation
Create a new file e.g. `foo.go` in the `internal/infrastructure/repository` directory to implement repository interface:
    - `fooRepo`: Repository struct
    - `NewFooRepo(db *gorm.DB) foo.Repository`: Constructor.
    - `WithTx(ctx context.Context) foo.Repository`: Transaction method.
    - `CreateFoo(ctx context.Context, entity *Foo) (*foo.Foo, error)`
    - `UpdateFoo(ctx context.Context, entity *Foo) error`
    - `DeleteFoo(ctx context.Context, entity *Foo) error`
    - `BulkCreate(ctx context.Context, entities []*Foo) error`
    - `CountFoo(ctx context.Context, filter *Filter) (int64, error)`
    - `GetFooList(ctx context.Context, filter *Filter) ([]*foo.Foo, error)`
    - `GetFooByID(ctx context.Context, id string) (*foo.Foo, error)`

## 4. Define the request and response DTOs
1. Create a new file e.g. `foo.go` in the `internal/delivery/http/dto/request` directory to define the request struct:
    - `FooCreateRequest`
    - `FooUpdateRequest`
    - `FooListRequest`
2. Create a new file e.g. `foo.go` in the `internal/delivery/http/dto/response` directory to define the response struct:
    - `FooResponse`


3. Create a new file e.g. `foo.go` in the `internal/delivery/http/request` directory to define the request converter method:
    - `func ToFooFilter(req *dtorequest.FooListRequest, ctx *fiber.Ctx) *foo.Filter`
4. Create a new file e.g. `foo.go` in the `internal/delivery/http/presenter` directory to define the response converter method:
    - `func ToFooResponse(entity *foo.Foo) *dtoresponse.FooResponse`
    - `func ToFooListResponse(entities []*foo.Foo) []*dtoresponse.FooResponse`

## 5. Define the Handler
Create a new file e.g. `foo.go` in the `internal/delivery/http/handler` directory, then define as follows:
    - `Foo`: Define the handler struct.
    - `NewFoo(validator *validator.Validate, usecase foo.Usecase) *Foo`: Define the handler constructor.   
    - `func(h *Foo) Create(ctx *fiber.Ctx) error`
    - `func(h *Foo) Update(ctx *fiber.Ctx) error`
    - `func(h *Foo) Delete(ctx *fiber.Ctx) error`
    - `func(h *Foo) GetList(ctx *fiber.Ctx) error`
    - `func(h *Foo) GetByID(ctx *fiber.Ctx) error`

## 6. Add Wire Binding
Add the wire binding in the `internal/wire/` directory.
1. Bind your repository implementation to the interface in the `internal/wire/repository.go` file.
2. Bind your usecase implementation to the interface in the `internal/wire/usecase.go` file.
3. Bind your handler implementation to the interface in the `internal/wire/handler.go` file.

## 7. Add permissions
Add permissoin constant in `pkg/constants/permission.go` file.

## 8. Add API Route
Register your API route:
1. If you need to define as public API, so register in `internal/delivery/http/router/public.go` file.
2. If you need to define as partner API, so register in `internal/delivery/http/router/partner.go` file.
3. If you need to define as internal API, so register in `internal/delivery/http/router/internal.go` file.

## 9. Attention 
Please note. If your API requires data from more than one domain, you must implement the application layer to serve as a multiple domain orchestrator.