# Baselix API Reference

This document describes the current public Baselix API contract.

## Base URL

```text
https://baselix.dev/api/v1
```

The public health check lives outside that prefix:

```text
https://baselix.dev/ping
```

## Authentication

All `/api/v1` routes require a project API key in the `Authorization` header.

```http
Authorization: Bearer <api_key>
```

If the header is missing, malformed, or the key is invalid, Baselix returns `401`.

## Common Error Shape

Most errors use this shape:

```json
{
  "error": 400,
  "message": "human-readable detail"
}
```

Current status codes you should expect:

| Status | Meaning |
| --- | --- |
| `400` | Invalid body, invalid UUID, reserved field, invalid filter or sort format, or duplicate unique value |
| `401` | Missing or invalid API key |
| `402` | Plan limit exceeded during record creation |
| `404` | Table or record not found |
| `500` | Internal error or an input path that currently bubbles up as a server error |

## Response Conventions

- Table list responses use `tables`.
- Record list responses use `records`.
- `GET /tables/{name}/{id}` also returns a `records` key, even though it contains a single object.
- Record objects are flat JSON objects with system fields plus your custom fields.

## Health Check

### `GET /ping`

Public health-check endpoint.

#### Request

```bash
curl https://baselix.dev/ping
```

#### Response `200`

```json
{
  "message": "pong"
}
```

## Tables

### `GET /tables`

List all tables for the authenticated project.

#### Request

```bash
curl https://baselix.dev/api/v1/tables \
  -H "Authorization: Bearer $BASELIX_API_KEY"
```

#### Response `200`

```json
{
  "tables": [
    {
      "name": "customers",
      "fields": {
        "email": "string_u",
        "age": "int",
        "is_active": "bool",
        "joined_at": "time",
        "profile": "json"
      }
    }
  ]
}
```

#### Errors

- `401` if the API key is missing or invalid
- `500` if Baselix cannot load the tables

### `GET /tables/{name}/schema`

Fetch the schema for one table.

#### Path parameters

| Name | Type | Description |
| --- | --- | --- |
| `name` | string | Table name |

#### Request

```bash
curl https://baselix.dev/api/v1/tables/customers/schema \
  -H "Authorization: Bearer $BASELIX_API_KEY"
```

#### Response `200`

```json
{
  "tables": {
    "name": "customers",
    "fields": {
      "email": "string_u",
      "age": "int",
      "is_active": "bool",
      "joined_at": "time",
      "profile": "json"
    }
  }
}
```

#### Errors

- `401` if the API key is missing or invalid
- `404` if the table does not exist
- `500` if Baselix cannot load the schema

### `POST /tables/{name}/schema`

Create a table with an explicit schema.

#### Path parameters

| Name | Type | Description |
| --- | --- | --- |
| `name` | string | Table name |

#### Request body

```json
{
  "schema": {
    "email": "string_u",
    "age": "int",
    "is_active": "bool",
    "joined_at": "time",
    "profile": "json"
  }
}
```

#### Request notes

- Supported types: `string`, `int`, `float`, `bool`, `time`, `json`, `uuid`
- Supported unique forms: `string_u`, `int_u`, `float_u`, `uuid_u`
- Reserved fields are not allowed: `id`, `created_at`, `updated_at`

#### Request example

```bash
curl -X POST https://baselix.dev/api/v1/tables/customers/schema \
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

#### Response `200`

```json
{
  "message": "Table 'customers' created successfully for project 'my-project'",
  "table_name": "customers"
}
```

#### Errors

- `400` for invalid JSON, reserved fields, or a table that already exists
- `401` if the API key is missing or invalid
- `500` for invalid schema type strings or other failures

### `PUT /tables/{name}/schema`

Replace or reconcile a table schema.

This route is destructive. If you remove a field or change a field type, existing data may be removed.

#### Path parameters

| Name | Type | Description |
| --- | --- | --- |
| `name` | string | Table name |

#### Request body

```json
{
  "schema": {
    "email": "string_u",
    "age": "int",
    "is_active": "bool",
    "tier": "string"
  },
  "allow_destructive": true
}
```

#### Request notes

- `allow_destructive` must be `true`
- Fields omitted from the new schema are removed
- New fields are added
- Changing a field type can remove existing values for that field

#### Request example

```bash
curl -X PUT https://baselix.dev/api/v1/tables/customers/schema \
  -H "Authorization: Bearer $BASELIX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "schema": {
      "email": "string_u",
      "age": "int",
      "is_active": "bool",
      "tier": "string"
    },
    "allow_destructive": true
  }'
```

#### Response `200`

```json
{
  "message": "Table 'customers' updated successfully for project 'my-project'",
  "table_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "table_name": "customers"
}
```

#### Errors

- `400` if `allow_destructive` is missing or false, or if reserved fields are used
- `401` if the API key is missing or invalid
- `500` for invalid schema type strings or other failures

### `DELETE /tables/{name}/schema`

Delete a table and all of its records.

#### Path parameters

| Name | Type | Description |
| --- | --- | --- |
| `name` | string | Table name |

#### Request body

```json
{
  "allow_destructive": true
}
```

#### Request example

```bash
curl -X DELETE https://baselix.dev/api/v1/tables/customers/schema \
  -H "Authorization: Bearer $BASELIX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "allow_destructive": true
  }'
```

#### Response `200`

```json
{
  "message": "Table 'customers' deleted successfully for project 'my-project'"
}
```

#### Errors

- `400` if the body is missing, invalid, or `allow_destructive` is false
- `401` if the API key is missing or invalid
- `404` if the table does not exist
- `500` for other delete failures

## Records

### `GET /tables/{name}`

List records from a table. Supports filtering and sorting.

#### Path parameters

| Name | Type | Description |
| --- | --- | --- |
| `name` | string | Table name |

#### Query parameters

| Name | Type | Description |
| --- | --- | --- |
| `filter` | string | Comma-separated filters in `field:operator:value` format |
| `sort` | string | Comma-separated sorts in `field:direction` format |

#### Supported filter operators

- `eq`
- `neq`
- `nq`
- `gt`
- `gte`
- `ge`
- `lt`
- `lte`
- `le`
- `contains`

#### Query examples

- `?filter=is_active:eq:true`
- `?filter=age:neq:21`
- `?filter=joined_at:gte:2026-04-01T00:00:00Z`
- `?filter=profile:contains:pro`
- `?sort=created_at:desc,email:asc`

#### Request example

```bash
curl "https://baselix.dev/api/v1/tables/customers?filter=joined_at:gte:2026-04-01T00:00:00Z&sort=created_at:desc,email:asc" \
  -H "Authorization: Bearer $BASELIX_API_KEY"
```

#### Response `200`

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
    }
  ]
}
```

#### Response notes

- The API can filter on `id`, `created_at`, and `updated_at` as well as your schema fields.
- `contains` works only on `string` and `json` fields.
- `field:eq:null` returns records where the field has no value.
- `field:neq:null` and `field:nq:null` return records where the field exists.
- The current API does not provide pagination.
- Record-list responses are currently capped at `1000` items.

#### Errors

- `400` for invalid filter or sort syntax, invalid sort fields, or unknown filter fields
- `401` if the API key is missing or invalid
- `404` if the table does not exist
- `500` for invalid value coercion or other failures

### `GET /tables/{name}/{id}`

Fetch one record by UUID.

#### Path parameters

| Name | Type | Description |
| --- | --- | --- |
| `name` | string | Table name |
| `id` | UUID string | Record ID |

#### Request example

```bash
curl https://baselix.dev/api/v1/tables/customers/3fa85f64-5717-4562-b3fc-2c963f66afa6 \
  -H "Authorization: Bearer $BASELIX_API_KEY"
```

#### Response `200`

```json
{
  "records": {
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
  }
}
```

#### Errors

- `400` if `id` is not a valid UUID
- `401` if the API key is missing or invalid
- `404` if the record does not exist

### `POST /tables/{name}`

Create one or more records.

#### Path parameters

| Name | Type | Description |
| --- | --- | --- |
| `name` | string | Table name |

#### Request body

Accepted shapes:

- One JSON object
- An array of JSON objects

Example single object:

```json
{
  "email": "ada@example.com",
  "age": 34,
  "is_active": true,
  "joined_at": "2026-04-01T10:00:00Z",
  "profile": {
    "plan": "pro"
  }
}
```

Example array:

```json
[
  {
    "email": "ada@example.com",
    "age": 34,
    "is_active": true
  },
  {
    "email": "grace@example.com",
    "age": 29,
    "is_active": false
  }
]
```

#### Request notes

- Reserved fields are not allowed: `id`, `created_at`, `updated_at`
- If the table does not exist, Baselix creates it automatically from the submitted payload
- Auto-created schemas do not infer unique fields
- A single request can currently include up to `1000` records
- Plan limits are enforced here

#### Request example

```bash
curl -X POST https://baselix.dev/api/v1/tables/customers \
  -H "Authorization: Bearer $BASELIX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '[
    {
      "email": "ada@example.com",
      "age": 34,
      "is_active": true,
      "joined_at": "2026-04-01T10:00:00Z",
      "profile": {
        "plan": "pro"
      }
    }
  ]'
```

#### Response `201`

If the table already exists:

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
        "plan": "pro"
      }
    }
  ]
}
```

If the table was auto-created:

```json
{
  "message": "Table 'customers' created with inferred schema.",
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
        "plan": "pro"
      }
    }
  ]
}
```

#### Errors

- `400` for empty bodies, invalid JSON, reserved fields, duplicate unique values, or too many records in the request
- `401` if the API key is missing or invalid
- `402` if the project has exceeded its plan limit
- `500` for type mismatches, bad inferred schema types, or other failures

#### Special-case error shape after table auto-creation

If Baselix creates the table and then record insertion fails, the response can include both `message` and `details`:

```json
{
  "message": "Table 'customers' created with inferred schema.",
  "error": 400,
  "details": "duplicate value: 'ada@example.com' for unique field: email"
}
```

### `PATCH /tables/{name}`

Update one or more records by ID.

#### Path parameters

| Name | Type | Description |
| --- | --- | --- |
| `name` | string | Table name |

#### Request body

Current accepted shapes:

- One JSON object with `id`
- An array of JSON objects with `id`

Recommended shape:

```json
[
  {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "is_active": false,
    "profile": {
      "plan": "enterprise"
    }
  }
]
```

#### Request notes

- `id` is required for every record update
- `created_at` and `updated_at` are not allowed in the payload
- `updated_at` is refreshed automatically
- Fields not present in the schema are rejected

#### Request example

```bash
curl -X PATCH https://baselix.dev/api/v1/tables/customers \
  -H "Authorization: Bearer $BASELIX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '[
    {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "is_active": false,
      "profile": {
        "plan": "enterprise"
      }
    }
  ]'
```

#### Response `200`

```json
{
  "records": [
    {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "created_at": "2026-04-14T12:00:00Z",
      "updated_at": "2026-04-14T13:15:00Z",
      "email": "ada@example.com",
      "age": 34,
      "is_active": false,
      "joined_at": "2026-04-01T10:00:00Z",
      "profile": {
        "plan": "enterprise"
      }
    }
  ]
}
```

#### Errors

- `400` for invalid JSON, missing `id`, invalid UUIDs inside the payload, reserved fields, or duplicate unique values
- `401` if the API key is missing or invalid
- `404` if the table or record does not exist
- `500` for unknown fields or other failures

### `DELETE /tables/{name}/{id}`

Delete one record by ID.

#### Path parameters

| Name | Type | Description |
| --- | --- | --- |
| `name` | string | Table name |
| `id` | UUID string | Record ID |

#### Request example

```bash
curl -X DELETE https://baselix.dev/api/v1/tables/customers/3fa85f64-5717-4562-b3fc-2c963f66afa6 \
  -H "Authorization: Bearer $BASELIX_API_KEY"
```

#### Response `200`

```json
{
  "message": "record deleted"
}
```

#### Errors

- `400` if `id` is not a valid UUID
- `401` if the API key is missing or invalid
- `404` if the record does not exist
- `500` for other delete failures

## Dashboard API-Key Lifecycle

Project API keys are not created through `/api/v1`.

Current workflow:

1. Sign in at `https://baselix.dev/sign-in`
2. Open the dashboard
3. Create a project to receive an API key
4. Rotate the key from the dashboard whenever needed

Treat the displayed key as a secret and replace old keys in your clients immediately after rotation.