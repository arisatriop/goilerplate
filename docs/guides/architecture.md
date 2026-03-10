# Clean Architecture Guide

In-depth explanation of Clean Architecture used in Goilerplate and how the layers work together.

---

## 🏛️ Architecture Overview

Goilerplate mengikuti **Clean Architecture** principles dengan clear separation of concerns:

```
┌─────────────────────────────────────────────────────┐
│                   DELIVERY LAYER                     │
│  ┌────────────┐  ┌──────────┐  ┌────────────────┐  │
│  │    DTO     │  │ Request  │  │    Presenter   │  │
│  │  (Structs) │  │ (Parsers)│  │  (Formatters)  │  │
│  └────────────┘  └──────────┘  └────────────────┘  │
│         ↓              ↓                  ↑          │
│  ┌──────────────────────────────────────────────┐  │
│  │             Handler (Orchestration)           │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
                        ↓↑
┌─────────────────────────────────────────────────────┐
│               APPLICATION LAYER                      │
│  ┌──────────────────────────────────────────────┐  │
│  │      Application Services (Use Case Orch.)   │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
                        ↓↑
┌─────────────────────────────────────────────────────┐
│                  DOMAIN LAYER                        │
│  ┌────────────┐  ┌──────────┐  ┌────────────────┐  │
│  │  Entities  │  │ Use Cases│  │   Interfaces   │  │
│  │ (Business) │  │ (Logic)  │  │  (Contracts)   │  │
│  └────────────┘  └──────────┘  └────────────────┘  │
└─────────────────────────────────────────────────────┘
                        ↓↑
┌─────────────────────────────────────────────────────┐
│             INFRASTRUCTURE LAYER                     │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────┐  │
│  │  Repository  │  │    Cache     │  │  Models │  │
│  │(Database Ops)│  │   (Redis)    │  │ (GORM)  │  │
│  └──────────────┘  └──────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## 📍 Layer Responsibilities

### 🖥️ Delivery Layer (HTTP)

**File location:** `internal/delivery/http/`

Responsible for handling HTTP requests and responses.

**Components:**

1. **DTOs (Data Transfer Objects)** - `dto/`
   - Pure data structures, no business logic
   - Validates struct tags
   - Used to serialize/deserialize HTTP data

   ```go
   type CreateUserRequest struct {
       Name     string `json:"name" validate:"required"`
       Email    string `json:"email" validate:"required,email"`
       Password string `json:"password" validate:"required,min=8"`
   }
   ```

2. **Request Parsers** - `request/`
   - Convert HTTP input → Domain objects
   - Extract query params, headers, body
   - Return domain-specific data structures

   ```go
   func ToUserFilter(req *dtoRequest.ListRequest, ctx *fiber.Ctx) *user.Filter {
       return &user.Filter{
           Search: req.Search,
           Page:   req.Page,
           Limit:  req.Limit,
       }
   }
   ```

3. **Presenters** - `presenter/`
   - Convert Domain entities → DTOs
   - Format complex data structures
   - Handle response transformation

   ```go
   func ToUserResponse(entity *user.User) *dtoResponse.UserResponse {
       return &dtoResponse.UserResponse{
           ID:    entity.ID,
           Name:  entity.Name,
           Email: entity.Email,
       }
   }
   ```

4. **Handlers** - `handler/`
   - Thin orchestration layer
   - Call application services/usecases
   - Coordinate request → response flow

   ```go
   func (h *User) Create(ctx *fiber.Ctx) error {
       // 1. Parse request
       req := &dtoRequest.CreateUserRequest{}
       if err := ctx.BodyParser(req); err != nil {
           return response.HandleError(ctx, err)
       }

       // 2. Execute business logic
       user, err := h.Usecase.Create(ctx.UserContext(), &user.User{...})

       // 3. Present response
       return response.Success(ctx, presenter.ToUserResponse(user))
   }
   ```

5. **Middleware** - `middleware/`
   - Cross-cutting concerns
   - Authentication, authorization, logging
   - Request validation

6. **Router** - `router/`
   - Define API routes
   - Organize routes (public, partner, internal)
   - Apply route-specific middleware

---

### 🔄 Request Flow Through Delivery Layer

```
HTTP Request arrives
     ↓
Middleware validates (auth, permissions, CORS)
     ↓
Handler receives request
     ↓
Request Parser transforms HTTP → Domain filter/command
     (example: request.ToUserFilter(ctx))
     ↓
Handler calls Application Service or Usecase
     (example: appService.GetUsers(ctx, filter))
     ↓
Data returned from Domain/Application layer
     ↓
Presenter transforms Domain → HTTP response
     (example: presenter.ToUserListResponse(entities))
     ↓
JSON response sent to client
```

---

### 💼 Application Layer

**File location:** `internal/application/`

Orchestrates use cases that involve multiple domains or complex business logic.

```go
type CreateOrderService struct {
    userUsecase    user.Usecase
    productUsecase product.Usecase
    orderUsecase   order.Usecase
}

func (s *CreateOrderService) Execute(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    // Validate user exists
    user, err := s.userUsecase.GetByID(ctx, req.UserID)

    // Validate products available
    products, err := s.productUsecase.GetByIDs(ctx, req.ProductIDs)

    // Create order (multi-domain orchestration)
    order, err := s.orderUsecase.Create(ctx, &order.Order{...})

    return order, nil
}
```

**When to use Application layer:**
- API requires data from multiple domains
- Business logic is complex & involves several steps
- Coordination between services

**When NOT to use Application layer:**
- Simple CRUD operations
- Single domain involved

---

### 🎯 Domain Layer

**File location:** `internal/domain/`

Core business logic and entities. **Dependencies must not point to other layers.**

**Components:**

1. **Entities** - Business objects dengan identitas dan lifecycle

   ```go
   type User struct {
       ID        string
       Name      string
       Email     string
       Password  string
       IsActive  bool
       CreatedAt time.Time
   }

   func (u *User) SetPassword(raw string) error {
       // Business logic for password hashing
       hash, err := bcrypt.GenerateFromPassword(...)
       u.Password = hash
       return nil
   }
   ```

2. **Use Cases** - Application rules & business logic

   ```go
   type Usecase interface {
       Create(ctx context.Context, entity *User) error
       GetByID(ctx context.Context, id string) (*User, error)
       GetList(ctx context.Context, filter *Filter) ([]*User, error)
       Update(ctx context.Context, entity *User) error
       Delete(ctx context.Context, id string) error
   }
   ```

3. **Repository Interfaces** - Contracts for data access

   ```go
   type Repository interface {
       CreateUser(ctx context.Context, entity *User) (*User, error)
       GetUserByID(ctx context.Context, id string) (*User, error)
       UpdateUser(ctx context.Context, entity *User) error
       DeleteUser(ctx context.Context, id string) error
   }
   ```

4. **Value Objects** - Immutable objects without identity

   ```go
   type Email struct {
       value string
   }

   func NewEmail(email string) (*Email, error) {
       if !isValidEmail(email) {
           return nil, ErrInvalidEmail
       }
       return &Email{value: email}, nil
   }
   ```

**Key Rule:** Domain layer does NOT know about HTTP, Database, or any Framework.

---

### 🗄️ Infrastructure Layer

**File location:** `internal/infrastructure/`

External integrations & implementation details.

**Components:**

1. **Models** - GORM database structs

   ```go
   type User struct {
       ID        string     `gorm:"primaryKey"`
       Name      string     `gorm:"column:name"`
       Email     string     `gorm:"column:email"`
       Password  string     `gorm:"column:password"`
       CreatedAt time.Time  `gorm:"column:created_at"`
       UpdatedAt time.Time  `gorm:"column:updated_at"`
   }
   ```

2. **Repository Implementations** - Concrete repository implementations

   ```go
   type userRepository struct {
       db *gorm.DB
   }

   func (r *userRepository) CreateUser(ctx context.Context, entity *domain.User) (*domain.User, error) {
       model := &User{
           ID:    entity.ID,
           Name:  entity.Name,
           Email: entity.Email,
       }
       if err := r.db.Create(model).Error; err != nil {
           return nil, err
       }
       return entity, nil
   }
   ```

3. **Cache** - Redis integration for caching & sessions

   ```go
   type SessionCache struct {
       client *redis.Client
   }

   func (c *SessionCache) Get(ctx context.Context, key string) (string, error) {
       return c.client.Get(ctx, key).Result()
   }
   ```

4. **Transactions** - Database transaction handling

   ```go
   func (r *userRepository) WithTx(ctx context.Context) Repository {
       tx := getTxFromContext(ctx).WithContext(ctx)
       return &userRepository{db: tx}
   }
   ```

---

## 🔀 Dependency Flow (Most Important!)

```
┌─────────────────┐
│  DELIVERY LAYER │ ← HTTP, REST specifics
└────────┬────────┘
         ↓ imports
┌─────────────────┐
│ APPLICATION     │ ← Use cases orchestration
│ LAYER           │
└────────┬────────┘
         ↓ imports
┌─────────────────┐
│  DOMAIN LAYER   │ ← Business logic (PURE)
└────────┬────────┘
         ↓ imports
┌─────────────────┐
│ INFRASTRUCTURE  │ ← Database, frameworks
│ LAYER           │
└─────────────────┘
```

**Key Rule:** Dependencies point **INWARD** only
- Delivery can import Application
- Application can import Domain
- Infrastructure implements Domain interfaces
- Domain **MUST NOT** import anyone

---

## 📚 Complete Request Example

**User performs:** `POST /api/v1/users`

### Step 1: HTTP Request arrives

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123"
}
```

### Step 2: Router → Middleware → Handler

```go
// router/public.go
router.Post("/users",
    middleware.Authenticate(),
    middleware.RequiredPermission("users.create"),
    handler.User.Create)

// handler/user.go
func (h *User) Create(ctx *fiber.Ctx) error {
    // Parse request
    req := &dtoRequest.CreateUserRequest{}
    ctx.BodyParser(req)

    // Call usecase
    user, err := h.Usecase.Create(ctx.UserContext(), &domain.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: req.Password,
    })

    // Present response
    response := presenter.ToUserResponse(user)
    return ctx.JSON(response)
}
```

### Step 3: Usecase (Domain Layer)

```go
// domain/user/usecase.go
func (u *usecase) Create(ctx context.Context, entity *User) error {
    // Business logic validation
    if !isValidEmail(entity.Email) {
        return ErrInvalidEmail
    }

    // Set password (business logic)
    entity.SetPassword(entity.Password)

    // Repository (abstraction, doesn't know about DB)
    return u.repo.CreateUser(ctx, entity)
}
```

### Step 4: Repository (Infrastructure Layer)

```go
// infrastructure/repository/user.go
func (r *userRepository) CreateUser(ctx context.Context, entity *domain.User) error {
    model := &model.User{
        ID:       entity.ID,
        Name:     entity.Name,
        Email:    entity.Email,
        Password: entity.Password,
    }
    return r.db.Create(model).Error
}
```

### Step 5: Database saves data

```sql
INSERT INTO users (id, name, email, password, created_at)
VALUES (...)
```

### Step 6: Response flows back

```
Database → Repository → Usecase → Handler → Presenter → JSON Response
```

---

## 🎯 Benefits of This Architecture

| Benefit | Description |
|---------|-------------|
| **Testability** | Each layer can be tested in isolation |
| **Maintainability** | Clear separation makes changes easier |
| **Reusability** | Presenters/parsers work with any delivery (gRPC, WebSocket) |
| **Scalability** | Easy to add new features without breaking existing |
| **Clean Code** | Handlers stay thin, logic stays in domain |
| **Type Safety** | Strong typing throughout all layers |
| **Framework Agnostic** | Domain logic is not bound to any framework |

---

## 💡 Common Patterns

### Pattern 1: Simple CRUD

For straightforward CRUD operations:

```
Handler → Usecase → Repository → Database
```

No need for Application layer or complex orchestration.

### Pattern 2: Multi-Domain Operation

For operations involving multiple domains:

```
Handler → Application Service → Multiple Usecases → Repositories → Database
```

Application Service orchestrates multiple domain usecases.

### Pattern 3: Complex Business Logic

For complex rules & validations:

```
Handler → Usecase → Domain Services → Repository
```

Use domain services for business logic helpers.

---

## 🔗 Related Documentation

- [CRUD Operations Guide](./crud-operations.md) - How to implement CRUD
- [Router & Routes](../api/router.md) - API organization
- [Development Guide](../getting-started/development.md) - Development workflow

---

## 📖 External Resources

- [Clean Architecture by Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Domain-Driven Design](https://www.domainlanguage.com/ddd/)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
