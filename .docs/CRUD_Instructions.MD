# CRUD Instructions 
Below are step by step to create new CRUD operations. You can use this instruction to create new CRUD operations for AI Agent.

## 1. Define the Domain
Create a new folder in the `internal/domain` directory with the name of your domain.
Inside the folder e.g `template`, create the following files:
1. `error.go`: Define the error types if needed (Optional).
2. `filter.go`: Define the filter struct if needed (Optional).
3. `message.go`: Define the message constants if needed (Optional).
4. `entity.go`: Define the entity struct.
5. `repository.go`: Define the repository interface.
    Inside this file, define the repository interface with the basic CRUD methods:
    - `WithTx(ctx context.Context) Repository`: Define the repository interface.
    - `CreateTemplate(ctx context.Context, entities *Template) (*Template, error)`: Define the repository interface.
    - `UpdateTemplate(ctx context.Context, entities *Template) error`: Define the repository interface.
    - `DeleteTemplate(ctx context.Context, entities *Template) error`: Define the repository interface.
    - `BulkCreate(ctx context.Context, entities []*Template) error`: Define the repository interface.
    - `CountTemplate(ctx context.Context, filter *Filter) (int64, error)`: Define the repository interface.
    - `GetTemplateList(ctx context.Context, filter *Filter) ([]*Template, error)`: Define the repository interface.
    - `GetTemplateByID(ctx context.Context, id string) (*Template, error)`: Define the repository interface.
6. `usecase.go`: Define the usecase interface with its implementation.
    Inside this file, define the usecase interface and its implementation with the basic CRUD methods:
    - `Create(ctx context.Context, entity *Template) error`: Define the usecase implementation.
    - `Update(ctx context.Context, entity *Template) error`: Define the usecase implementation.
    - `Delete(ctx context.Context, entity *Template) error`: Define the usecase implementation.
    - `BulkCreate(ctx context.Context, entities []*Template) error`: Define the usecase implementation.
    - `Count(ctx context.Context, filter *Filter) (int64, error)`: Define the usecase implementation.
    - `GetList(ctx context.Context, filter *Filter) ([]*Template, error)`: Define the usecase implementation.
    - `GetByID(ctx context.Context, id string) (*Template, error)`: Define the usecase implementation.

## 2. Define the model 
Create a new file e.g `template.go` in the `internal/infrastructure/model` directory, then define the struct with basic field as follow:
```
    type Template struct {
	ID        string     `gorm:"primaryKey;default:gen_random_uuid()"`
	Code      string     `gorm:"column:code"`
	Template  string     `gorm:"column:template"`
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
Create a new file e.g `template.go` in the `internal/infrastructure/repository` directory, then define as follows:
    - `templateRepo`: Define the repository implementation.
    - `NewTemplateRepo(db *gorm.DB) template.Repository`: Define the repository constructor.
    - `WithTx(ctx context.Context) template.Repository`: Define the repository implementation.
    - `CreateTemplate(ctx context.Context, entity *Template) (*template.Template, error)`: Define the repository implementation.
    - `UpdateTemplate(ctx context.Context, entity *Template) error`: Define the repository implementation.
    - `DeleteTemplate(ctx context.Context, entity *Template) error`: Define the repository implementation.
    - `BulkCreate(ctx context.Context, entities []*Template) error`: Define the repository implementation.
    - `CountTemplate(ctx context.Context, filter *Filter) (int64, error)`: Define the repository implementation.
    - `GetTemplateList(ctx context.Context, filter *Filter) ([]*template.Template, error)`: Define the repository implementation.
    - `GetTemplateByID(ctx context.Context, id string) (*template.Template, error)`: Define the repository implementation.

## 4. Define the request and response DTOs
1. Create a new file e.g `template.go` in the `internal/delivery/http/dto/request` directory, then define the struct:
    - `TemplateCreateRequest`: Define the request DTOs.
    - `TemplateUpdateRequest`: Define the request DTOs.
    - `TemplateListRequest`: Define the request DTOs.
2. Create a new file e.g `template.go` in the `internal/delivery/http/dto/response` directory, then define the struct:
    - `TemplateResponse`: Define the response DTOs.

## 5. Define the Handler
Create a new file e.g `template.go` in the `internal/delivery/http/handler` directory, then define as follows:
    - `Template`: Define the handler struct.
    - `NewTemplate(validator *validator.Validate, usecase template.Usecase) *Template`: Define the handler constructor.   
    - `func(h *Template) Create(ctx *fiber.Ctx) error`: Define the handler implementation.
    - `func(h *Template) Update(ctx *fiber.Ctx) error`: Define the handler implementation.
    - `func(h *Template) Delete(ctx *fiber.Ctx) error`: Define the handler implementation.
    - `func(h *Template) GetList(ctx *fiber.Ctx) error`: Define the handler implementation.
    - `func(h *Template) GetByID(ctx *fiber.Ctx) error`: Define the handler implementation.

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

## 9. PLEASE ATTENTION 
Please note. If your API requires data from more than one domain, you must implement the application layer to serve as a multiple domain orchestrator.