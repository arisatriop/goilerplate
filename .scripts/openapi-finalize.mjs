// Makes the converted document tool-neutral, and adds what its Swagger 2.0 source cannot express.
// Run by `make openapi`.
import { readFileSync, writeFileSync } from 'node:fs';

const path = process.argv[2];
let text = readFileSync(path, 'utf8');

// Schema names arrive as Go package.Type. Dots are legal, but they leak the code's layout into the
// contract and make client generators produce awkward names, so publish the bare type name.
const { schemas } = JSON.parse(text).components;
const renames = new Map();
for (const name of Object.keys(schemas)) {
  const bare = name.slice(name.lastIndexOf('.') + 1);
  const published = bare[0].toUpperCase() + bare.slice(1);
  if ([...renames.values()].includes(published)) {
    throw new Error(`two schemas would both be published as ${published}`);
  }
  renames.set(name, published);
}
for (const [from, to] of renames) {
  text = text.replaceAll(`"#/components/schemas/${from}"`, `"#/components/schemas/${to}"`);
}

const spec = JSON.parse(text);
spec.components.schemas = Object.fromEntries(
  Object.entries(spec.components.schemas)
    .map(([name, schema]) => [renames.get(name), schema])
    .sort(([a], [b]) => a.localeCompare(b)),
);

// 3.0.3 is the newest 3.0 release, only corrections over 3.0.0, and the version tools most widely
// support.
spec.openapi = '3.0.3';
delete spec.info.contact;

// Swagger 2.0 has no bearer scheme, so the source can only declare an API key named Authorization.
// As a real bearer scheme, clients add the "Bearer " prefix themselves.
const { description } = spec.components.securitySchemes.BearerAuth;
spec.components.securitySchemes.BearerAuth = { type: 'http', scheme: 'bearer', bearerFormat: 'JWT', description };

// An operation with no security key inherits the top-level one, which this document does not set.
// Saying so explicitly is how OpenAPI marks an operation public, and strict validators require it.
for (const methods of Object.values(spec.paths)) {
  for (const operation of Object.values(methods)) {
    operation.security ??= [];
  }
}

// The converter keeps the base path's slash, which would double up in front of every path.
spec.servers = [{ url: 'http://localhost:3000', description: 'Local development' }];

writeFileSync(path, `${JSON.stringify(spec, null, 2)}\n`);
