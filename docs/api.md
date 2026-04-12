# Baselix API Reference

Baselix provides a dynamic data-storage REST API built on an Entity–Attribute–Value (EAV) model.
You define tables, manage their schemas, and perform full CRUD on records — no database migrations required.

A machine-readable [OpenAPI 3.0 spec](./openapi.yaml) is available for use with Swagger UI, Postman, and similar tools.

---

## Base URL

```
https://your-domain.com/api/v1
```

---

## Authentication

All `/api/v1` endpoints require a **project API key** passed as a Bearer token:

```
Authorization: Bearer <api_key>
```

API keys are generated per project from the Baselix dashboard and can be rotated at any time. An invalid or missing key returns:

```json
{ "error": 401, "message": "invalid API key" }
```

---

## Field Types

When defining a table schema, each field is mapped to a type string:

| Type     | Description          |
| -------- | -------------------- |
| `string` | Text                 |
| `int`    | Integer              |
| `float`  | Decimal number       |
| `bool`   | Boolean              |
| `time`   | ISO 8601 timestamp   |
| `json`   | JSON object or array |
| `uuid`   | UUID string          |

Append `_u` to enforce a **uniqueness constraint** on a field (e.g. `"email": "string_u"`).
The unique suffix is supported on `string`, `int`, `float`, and `uuid`.

The fields `id`, `created_at`, and `updated_at` are **reserved** and managed automatically.
Including them in a schema or record payload will return a `400` error.

---

## Filtering & Sorting

### Filter query parameter

```
?filter=field:operator:value
```

Multiple conditions are comma-separated:

```
?filter=age:gte:18,status:eq:active
```

**Supported operators:**

| Operator   | Meaning                                         |
| ---------- | ----------------------------------------------- |
| `eq`       | Equal                                           |
| `ne`       | Not equal                                       |
| `gt`       | Greater than                                    |
| `gte`      | Greater than or equal                           |
| `lt`       | Less than                                       |
| `lte`      | Less than or equal                              |
| `contains` | Case-insensitive substring match (strings only) |

You can filter on user-defined fields as well as the meta fields `id`, `created_at`, and `updated_at`.
Use `null` as the value for null checks (e.g. `?filter=email:eq:null`).

### Sort query parameter

```
?sort=field:direction
```

Multiple sorts are comma-separated:

```
?sort=created_at:desc,name:asc
```

`direction` must be `asc` or `desc`. Defaults to `asc` if unrecognized.

---

## Error Format

All errors follow the same shape:

```json
{
  "error": <http_status_code>,
  "message": "<description>"
}
```

**Common errors:**

| Status | Cause                                                                                         |
| ------ | --------------------------------------------------------------------------------------------- |
| `400`  | Invalid request body, reserved field name, duplicate unique value, invalid filter/sort format |
| `401`  | Missing or invalid API key                                                                    |
| `404`  | Table or record not found                                                                     |
| `500`  | Internal server error                                                                         |

---

## Endpoints

### Tables

Tables define the schema (shape) of your data. Each project can have multiple tables.

---

#### `GET /tables`

List all tables and their schemas for the authenticated project.

**Response `200`**

```json
{
  "tables": [
    {
      "name": "users",
      "fields": {
        "name": "string_u",
        "age": "int"
      }
    },
    {
      "name": "products",
      "fields": {
        "title": "string",
        "price": "float"
      }
    }
  ]
}
```

---

#### `GET /tables/:name/schema`

Get the schema for a single table.

**Path parameters**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `name`    | string | Table name  |

**Response `200`**

```json
{
  "tables": {
    "name": "users",
    "fields": {
      "name": "string_u",
      "age": "int"
    }
  }
}
```

**Response `404`** — table does not exist.

---

#### `POST /tables/:name/schema`

Create a new table with the given schema.

**Path parameters**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `name`    | string | Table name  |

**Request body**

```json
{
  "schema": {
    "name": "string_u",
    "age": "int",
    "score": "float"
  }
}
```

**Response `200`**

```json
{
  "message": "Table 'users' created successfully for project 'my-project'",
  "table_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "table_name": "users"
}
```

**Response `400`** — table already exists, or a reserved field name was used.

---

#### `PUT /tables/:name/schema`

Update the schema for an existing table.

> **Warning — Destructive operation.**
>
> - Fields removed from the new schema will be **deleted along with all their stored values**.
> - Fields whose type changes will have all their **existing values deleted**.
> - New fields are added to the schema; existing records will have `null` for those fields.
>
> You must explicitly acknowledge this by setting `"allow_destructive": true`.

**Path parameters**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `name`    | string | Table name  |

**Request body**

```json
{
  "schema": {
    "name": "string_u",
    "age": "int",
    "email": "string_u"
  },
  "allow_destructive": true
}
```

**Response `200`**

```json
{
  "message": "Table 'users' updated successfully for project 'my-project'",
  "table_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "table_name": "users"
}
```

**Response `400`** — `allow_destructive` is `false` or missing, reserved field name, or invalid schema.

---

#### `DELETE /tables/:name/schema`

Permanently delete a table and **all** of its data (fields, records, and values).

**Path parameters**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `name`    | string | Table name  |

**Response `200`**

```json
{
  "message": "Table 'users' deleted successfully for project 'my-project'"
}
```

**Response `404`** — table does not exist.

---

### Records

Records are the rows of your tables. Each record is identified by a UUID and includes system-managed
`id`, `created_at`, and `updated_at` fields alongside your user-defined fields.

---

#### `GET /tables/:name`

List all records in a table. Supports optional filtering and sorting via query parameters.

**Path parameters**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `name`    | string | Table name  |

**Query parameters**

| Parameter | Description                                                 | Example                          |
| --------- | ----------------------------------------------------------- | -------------------------------- |
| `filter`  | Filter conditions (`field:operator:value`, comma-separated) | `age:gte:18,name:contains:alice` |
| `sort`    | Sort directives (`field:direction`, comma-separated)        | `created_at:desc,name:asc`       |

**Response `200`**

```json
{
  "records": [
    {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "created_at": "2026-04-12T10:00:00Z",
      "updated_at": "2026-04-12T10:00:00Z",
      "name": "Alice",
      "age": 30
    }
  ]
}
```

**Response `400`** — invalid filter/sort format, or filter references a field not in the schema.

**Response `404`** — table not found.

---

#### `GET /tables/:name/:id`

Fetch a single record by its UUID.

**Path parameters**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `name`    | string | Table name  |
| `id`      | UUID   | Record ID   |

**Response `200`**

```json
{
  "records": {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "created_at": "2026-04-12T10:00:00Z",
    "updated_at": "2026-04-12T10:00:00Z",
    "name": "Alice",
    "age": 30
  }
}
```

**Response `400`** — `id` is not a valid UUID.

**Response `404`** — record not found.

---

#### `POST /tables/:name`

Create one or more records. Accepts a **single JSON object** or an **array of objects**.

> **Auto-schema inference:** If the table does not yet exist, it will be created automatically
> using a schema inferred from the first batch of records. When this happens, the response
> includes a `message` field.

**Path parameters**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `name`    | string | Table name  |

**Request body — single record**

```json
{
  "name": "Alice",
  "age": 30,
  "is_active": true
}
```

**Request body — multiple records**

```json
[
  { "name": "Alice", "age": 30 },
  { "name": "Bob", "age": 25 }
]
```

**Response `201` — inserted into existing table**

```json
{
  "records": [
    {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "created_at": "2026-04-12T10:00:00Z",
      "updated_at": "2026-04-12T10:00:00Z",
      "name": "Alice",
      "age": 30
    }
  ]
}
```

**Response `201` — table auto-created**

```json
{
  "message": "Table 'users' created with inferred schema.",
  "records": [
    {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "created_at": "2026-04-12T10:00:00Z",
      "updated_at": "2026-04-12T10:00:00Z",
      "name": "Alice",
      "age": 30
    }
  ]
}
```

**Response `400`** — reserved field included, or duplicate value for a unique field.

---

#### `PATCH /tables/:name`

Partially update one or more records. The body must be an **array** — even for a single update.
Each element must include an `id` (UUID) to identify the record, plus any fields to change.

`updated_at` is refreshed automatically. `created_at` and `updated_at` cannot be set manually.

**Path parameters**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `name`    | string | Table name  |

**Request body**

```json
[
  {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "age": 31,
    "is_active": false
  }
]
```

**Response `200`**

```json
{
  "records": [
    {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "created_at": "2026-04-12T09:00:00Z",
      "updated_at": "2026-04-12T11:00:00Z",
      "name": "Alice",
      "age": 31,
      "is_active": false
    }
  ]
}
```

**Response `400`** — reserved field included, or duplicate value for a unique field.

**Response `404`** — table or record not found.

---

#### `DELETE /tables/:name/:id`

Permanently delete a single record and all its stored values.

**Path parameters**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `name`    | string | Table name  |
| `id`      | UUID   | Record ID   |

**Response `200`**

```json
{
  "message": "record deleted"
}
```

**Response `400`** — `id` is not a valid UUID.

**Response `404`** — record not found.
