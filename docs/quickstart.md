# Baselix Quickstart

This guide walks through the shortest path from a new project to working API calls.

## 1. Get a Project API Key

Project API keys are created from the signed-in dashboard.

1. Go to `https://baselix.dev/sign-in` and sign in.
2. Open the dashboard.
3. Create a project.
4. Copy the generated API key.

If you rotate a key later, keep the newly displayed key immediately.

## 2. Set Local Variables

Examples below use the Baselix production URL.

### PowerShell

```powershell
$env:BASELIX_BASE_URL = "https://baselix.dev"
$env:BASELIX_API_KEY = "sk_your_project_key"
```

### POSIX shell

```bash
export BASELIX_BASE_URL="https://baselix.dev"
export BASELIX_API_KEY="sk_your_project_key"
```

## 3. Check Connectivity

`/ping` is public and does not require authentication.

```bash
curl "$BASELIX_BASE_URL/ping"
```

Expected response:

```json
{
  "message": "pong"
}
```

## 4. Create a Table Schema

Use the schema endpoint when you want exact field types and explicit uniqueness constraints.

```bash
curl -X POST "$BASELIX_BASE_URL/api/v1/tables/customers/schema" \
  -H "Authorization: Bearer $BASELIX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "schema": {
      "email": "string_u",
      "age": "int",
      "is_active": "bool",
      "joined_at": "time",
      "profile": "json"
    }
  }'
```

Example response:

```json
{
  "message": "Table 'customers' created successfully for project 'my-project'",
  "table_name": "customers"
}
```

## 5. Insert Records

Record creation accepts either a single JSON object or an array of objects.

```bash
curl -X POST "$BASELIX_BASE_URL/api/v1/tables/customers" \
  -H "Authorization: Bearer $BASELIX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '[
    {
      "email": "ada@example.com",
      "age": 34,
      "is_active": true,
      "joined_at": "2026-04-01T10:00:00Z",
      "profile": {
        "plan": "pro",
        "region": "eu-west-1"
      }
    },
    {
      "email": "grace@example.com",
      "age": 29,
      "is_active": false,
      "joined_at": "2026-04-03T15:30:00Z",
      "profile": {
        "plan": "free",
        "region": "us-east-1"
      }
    }
  ]'
```

Example response:

```json
{
  "records": [
    {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "created_at": "2026-04-14T12:00:00Z",
      "updated_at": "2026-04-14T12:00:00Z",
      "email": "ada@example.com",
      "age": 34,
      "is_active": true,
      "joined_at": "2026-04-01T10:00:00Z",
      "profile": {
        "plan": "pro",
        "region": "eu-west-1"
      }
    },
    {
      "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
      "created_at": "2026-04-14T12:00:00Z",
      "updated_at": "2026-04-14T12:00:00Z",
      "email": "grace@example.com",
      "age": 29,
      "is_active": false,
      "joined_at": "2026-04-03T15:30:00Z",
      "profile": {
        "plan": "free",
        "region": "us-east-1"
      }
    }
  ]
}
```

## 6. Query Records

Filtering and sorting are both optional query parameters.

```bash
curl "$BASELIX_BASE_URL/api/v1/tables/customers?filter=joined_at:gte:2026-04-01T00:00:00Z&sort=created_at:desc,email:asc" \
  -H "Authorization: Bearer $BASELIX_API_KEY"
```

Notes:

- Multiple filters are comma-separated and combined with AND.
- Multiple sorts are comma-separated and applied as a stable multi-sort.
- The `contains` operator works only on `string` and `json` fields.
- Not-equal filters use `neq` or `nq`.
- Greater-than-or-equal filters can use `gte` or `ge`.
- Less-than-or-equal filters can use `lte` or `le`.

## 7. Update Records

`PATCH` is intended for partial updates. The safest payload shape is an array of objects with `id` plus the fields you want to change.

```bash
curl -X PATCH "$BASELIX_BASE_URL/api/v1/tables/customers" \
  -H "Authorization: Bearer $BASELIX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '[
    {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "is_active": false,
      "profile": {
        "plan": "enterprise",
        "region": "eu-west-1"
      }
    }
  ]'
```

`updated_at` is refreshed automatically. You cannot set `created_at` or `updated_at` yourself.

## 8. Fetch or Delete a Single Record

### Fetch by ID

```bash
curl "$BASELIX_BASE_URL/api/v1/tables/customers/3fa85f64-5717-4562-b3fc-2c963f66afa6" \
  -H "Authorization: Bearer $BASELIX_API_KEY"
```

### Delete by ID

```bash
curl -X DELETE "$BASELIX_BASE_URL/api/v1/tables/customers/3fa85f64-5717-4562-b3fc-2c963f66afa6" \
  -H "Authorization: Bearer $BASELIX_API_KEY"
```

## 9. Delete a Table

Table deletion is destructive and requires an explicit body.

```bash
curl -X DELETE "$BASELIX_BASE_URL/api/v1/tables/customers/schema" \
  -H "Authorization: Bearer $BASELIX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "allow_destructive": true
  }'
```

## 10. Alternative: Let Baselix Infer the Schema

If you `POST` records to a table that does not exist yet, Baselix creates the table automatically by inferring field types from the submitted records.

That is convenient, but keep these tradeoffs in mind:

- Numeric fields infer as `float`, not `int`.
- Unique constraints are not inferred.
- Strings that parse as UUIDs or RFC3339 timestamps infer as `uuid` or `time`.
- List values are not supported during inference.
- Mixed value types across the same field can produce a schema you did not intend.

For production use, explicit schema creation is usually the safer starting point.

## JavaScript Example

```javascript
const BASE_URL = process.env.BASELIX_BASE_URL ?? "https://baselix.dev";
const API_KEY = process.env.BASELIX_API_KEY;

async function baselix(path, init = {}) {
  const response = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${API_KEY}`,
      ...(init.headers ?? {}),
    },
  });

  const body = await response.json();
  if (!response.ok) {
    throw new Error(JSON.stringify(body));
  }

  return body;
}

const result = await baselix("/api/v1/tables/customers?sort=created_at:desc");
console.log(result.records);
```

For the full route-by-route contract, continue to [API Reference](./api.md).
