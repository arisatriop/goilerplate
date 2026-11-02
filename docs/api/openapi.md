# OpenAPI Spec

[`docs/api/openapi.json`](./openapi.json) describes every HTTP route as a standard OpenAPI 3.0.3
document: the end-user API, `/internal`, `/partner/v1`, and the health and metrics endpoints. It
contains no tool-specific extensions, so it works with anything that reads OpenAPI: API clients
such as Postman, Apidog, Insomnia or Bruno, documentation renderers, mock servers, and client
generators.

gRPC is not part of it. Its contract is the
[goilerplate-proto](https://github.com/arisatriop/goilerplate-proto) module.

## Using it

Import the file into your tool. Tools that import from a URL can use
`http://localhost:3000/openapi.json` with the server running. That route exists only when `app.env`
is not `production`, and it reads the file from disk, so it needs the server started from the
repository root. The file itself works everywhere.

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

No migration grants those permissions on purpose. Registration makes every new account an `owner`,
so granting them to that role would let anyone who signs up change shared data. To try the bars
routes on a local database:

```sql
INSERT INTO permissions (id, name, slug, created_at, created_by, updated_at, updated_by)
SELECT gen_random_uuid(), slug, slug, now(), 'local', now(), 'local'
  FROM unnest(ARRAY['bar.list', 'bar.get', 'bar.create', 'bar.update', 'bar.delete']) AS slug;

INSERT INTO role_permissions (role_id, permission_id, created_at, created_by)
SELECT r.id, p.id, now(), 'local'
  FROM roles r CROSS JOIN permissions p
 WHERE r.slug = 'owner' AND p.slug LIKE 'bar.%';
```

With Redis, permissions are cached, and a direct SQL change does not invalidate the cache. It takes
effect within `auth.permission_cache_ttl`.

## Changing it

The spec is generated from the annotations on the handlers, so don't edit the file by hand. After
changing a route, a DTO or a status code:

```bash
make openapi    # needs Go and Node.js
```

Commit `docs/api/openapi.json` with the change. How the file gets built is an internal detail that
consumers never see: `make openapi` reads the annotations with
[swag](https://github.com/swaggo/swag), upgrades the result to OpenAPI 3 with
[swagger2openapi](https://github.com/Mermade/oas-kit), and `.scripts/openapi-finalize.mjs` removes
what is specific to Go or to the older format.

The bar routes share one handler across audiences, and an annotated function has one set of
security rules, so `/internal` and `/partner/v1` are annotated in
`internal/delivery/http/handler/bar_audiences.go`.

`TestOpenAPI_MatchesRegisteredRoutes` fails when a route is missing from the spec or the spec names
a route that does not exist. It checks routes only, so a changed DTO needs `make openapi` too.
