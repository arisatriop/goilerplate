// Corrects what the Swagger 2.0 to OpenAPI 3 conversion cannot express. Run by `make swag`.
import { readFileSync, writeFileSync } from 'node:fs';

const path = process.argv[2];
const spec = JSON.parse(readFileSync(path, 'utf8'));

// Swagger 2.0 has no bearer scheme, so swag can only declare an API key named Authorization.
// As a real bearer scheme, Postman and Apidog add the "Bearer " prefix themselves.
const { description } = spec.components.securitySchemes.BearerAuth;
spec.components.securitySchemes.BearerAuth = { type: 'http', scheme: 'bearer', bearerFormat: 'JWT', description };

// The converter keeps the base path's slash, which would double up in front of every path.
spec.servers = [{ url: 'http://localhost:3000', description: 'Local development' }];

writeFileSync(path, `${JSON.stringify(spec, null, 2)}\n`);
