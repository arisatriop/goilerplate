# CRUD Instructions 
Below are step by step to create new CRUD operations. You can use this instruction to create new CRUD operations for AI Agent.

## 1. Define the Domain
Create a new folder in the `internal/domain` directory with the name of your domain.
Inside the folder e.g `example`, create the following files:
1. `error.go`: Define the error types if needed (Optional).
2. `filter.go`: Define the filter struct if needed (Optional).
3. `message.go`: Define the message constants if needed (Optional).
4. `entity.go`: Define the entity struct.
5. `repository.go`: Define the repository interface.
    Inside this file, define the repository interface with the basic CRUD methods:
    - `WithTx(ctx context.Context) Repository`: Define the repository interface.
    - `CreateExample(ctx context.Context, entities *Example) (*Example, error)`: Define the repository interface.
    - `UpdateExample(ctx context.Context, entities *Example) error`: Define the repository interface.
    - `DeleteExample(ctx context.Context, entities *Example) error`: Define the repository interface.
    - `BulkCreate(ctx context.Context, entities []*Example) error`: Define the repository interface.
    - `CountExample(ctx context.Context, filter *Filter) (int64, error)`: Define the repository interface.
    - `GetExampleList(ctx context.Context, filter *Filter) ([]*Example, error)`: Define the repository interface.
    - `GetExampleByID(ctx context.Context, id string) (*Example, error)`: Define the repository interface.
6. `usecase.go`: Define the usecase interface with its implementation.
    Inside this file, define the usecase interface and its implementation with the basic CRUD methods:
    - `Create(ctx context.Context, entity *Example) error`: Define the usecase implementation.
    - `Update(ctx context.Context, entity *Example) error`: Define the usecase implementation.
    - `Delete(ctx context.Context, entity *Example) error`: Define the usecase implementation.
    - `BulkCreate(ctx context.Context, entities []*Example) error`: Define the usecase implementation.
    - `Count(ctx context.Context, filter *Filter) (int64, error)`: Define the usecase implementation.
    - `GetList(ctx context.Context, filter *Filter) ([]*Example, error)`: Define the usecase implementation.
    - `GetByID(ctx context.Context, id string) (*Example, error)`: Define the usecase implementation.

## 2. Define the repository implementation
Create a new file e.g `example.go` in the `internal/infrastructure/repository` directory, then define as follows:
    - `exampleRepo`: Define the repository implementation.
    - `NewExampleRepo(db *gorm.DB) example.Repository`: Define the repository constructor.
    - `WithTx(ctx context.Context) example.Repository`: Define the repository implementation.
    - `CreateExample(ctx context.Context, entity *Example) (*example.Example, error)`: Define the repository implementation.
    - `UpdateExample(ctx context.Context, entity *Example) error`: Define the repository implementation.
    - `DeleteExample(ctx context.Context, entity *Example) error`: Define the repository implementation.
    - `BulkCreate(ctx context.Context, entities []*Example) error`: Define the repository implementation.
    - `CountExample(ctx context.Context, filter *Filter) (int64, error)`: Define the repository implementation.
    - `GetExampleList(ctx context.Context, filter *Filter) ([]*example.Example, error)`: Define the repository implementation.
    - `GetExampleByID(ctx context.Context, id string) (*example.Example, error)`: Define the repository implementation.

## 3. Define the request and response DTOs
1. Create a new file e.g `example.go` in the `internal/delivery/http/dto/request` directory, then define the struct:
    - `ExampleCreateRequest`: Define the request DTOs.
    - `ExampleUpdateRequest`: Define the request DTOs.
    - `ExampleGetListRequest`: Define the request DTOs.
2. Create a new file e.g `example.go` in the `internal/delivery/http/dto/response` directory, then define the struct:
    - `ExampleResponse`: Define the response DTOs.

## 4. Define the Handler
Create a new file e.g `example.go` in the `internal/delivery/http/handler` directory, then define as follows:
    - `Example`: Define the handler struct.
    - `NewExample(validator *validator.Validate, usecase example.Usecase) *Example`: Define the handler constructor.   
    - `func(h *Example) Create(ctx *fiber.Ctx) error`: Define the handler implementation.
    - `func(h *Example) Update(ctx *fiber.Ctx) error`: Define the handler implementation.
    - `func(h *Example) Delete(ctx *fiber.Ctx) error`: Define the handler implementation.
    - `func(h *Example) GetList(ctx *fiber.Ctx) error`: Define the handler implementation.
    - `func(h *Example) GetByID(ctx *fiber.Ctx) error`: Define the handler implementation.

## 5. Add Wire Binding
Add the wire binding in the `internal/wire/` directory.
1. Bind your repository implementation to the interface in the `internal/wire/repository.go` file.
2. Bind your usecase implementation to the interface in the `internal/wire/usecase.go` file.
3. Bind your handler implementation to the interface in the `internal/wire/handler.go` file.

## 6. Add API Route
Register your API route:
1. If you need to define as public API, so register in `internal/delivery/http/router/public.go` file.
3. If you need to define as partner API, so register in `internal/delivery/http/router/partner.go` file.
4. If you need to define as internal API, so register in `internal/delivery/http/router/internal.go` file.