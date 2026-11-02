# OpenAPI Spec

[`.swagger/openapi.json`](../../.swagger/openapi.json) describes every HTTP route as OpenAPI 3.0:
the end-user API, `/internal`, `/partner/v1`, and the health and metrics endpoints.

gRPC is not part of it. Its contract is the
[goilerplate-proto](https://github.com/arisatriop/goilerplate-proto) module.

## Importing it

| Tool | How |
|---|---|
| Postman | **Import**, then choose `.swagger/openapi.json` |
| Apidog | **Settings → Import Data → OpenAPI/Swagger**, then choose the file |
| Swagger UI | Open `http://localhost:3000/swaggerui/index.html` with the server running |

Tools that import from a URL can use `http://localhost:3000/swagger-ui/openapi.json`. Both URLs
exist only when `app.env` is not `production`; the file works everywhere.

## Credentials

| Scheme | Used by | Value |
|---|---|---|
| `BearerAuth` | `/api/v1` | The `accessToken` from login. On `/api/v1/auth/refresh`, the `refreshToken` |
| `RefreshCookie` | `/api/v1/auth/refresh` | The refresh cookie, when `auth.refresh_transport=cookie` |
| `ApiKeyAuth` | `/partner/v1` | A key from `api_key.<name>` in the configuration |
| `InternalSecret` | `/internal` | `internal_auth.secret`, needed only when `internal_auth.mode=shared_secret` |

## What depends on the configuration

The spec lists every route, but a running server may not have them all:

- The email routes exist only with `auth.email.enabled`
- `/partner/v1` exists only when an API key is configured
- `/metrics` exists only with OpenTelemetry enabled
- `/api/v1/bars` answers 403 until a role holds the `bar.*` permissions. No migration grants them

## Changing it

The spec is generated from the annotations on the handlers, so don't edit the file by hand. After
changing a route, a DTO or a status code:

```bash
make swag    # needs swag v1.16.4 and Node.js
```

Commit `.swagger/openapi.json` with the change. `make swag` writes Swagger 2.0 with
[swag](https://github.com/swaggo/swag), converts it with
[swagger2openapi](https://github.com/Mermade/oas-kit), and then
`.scripts/openapi-finalize.mjs` sets what Swagger 2.0 cannot express.

The bar routes share one handler across audiences, and an annotated function has one set of
security rules, so `/internal` and `/partner/v1` are annotated in
`internal/delivery/http/handler/bar_audiences.go`.

`TestOpenAPI_MatchesRegisteredRoutes` fails when a route is missing from the spec or the spec names
a route that does not exist. It checks routes only, so a changed DTO needs `make swag` too.
