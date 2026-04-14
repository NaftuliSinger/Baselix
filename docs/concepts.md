# Concepts and Caveats

This page explains the public API model, type system, filtering rules, limits, and the runtime behaviors that matter when building against Baselix.

## Data Model

At the API level, Baselix gives you a simple model:

- A **project** owns tables and is selected by the API key you send.
- A **table** defines a schema.
- A **record** is one row in that table.
- Each record has system fields plus your custom fields.

## Authentication and Project Scope

Every `/api/v1` route requires:

```http
Authorization: Bearer <api_key>
```

The key selects a single project. All table and record operations are scoped to that project.

## Field Types

Schemas use string type names.

| Type | Meaning | Notes |
| --- | --- | --- |
| `string` | Text | Plain string storage |
| `int` | Integer | Whole numbers only |
| `float` | Decimal number | JSON numbers infer as this during auto-creation |
| `bool` | Boolean | `true` or `false` |
| `time` | RFC3339 timestamp | Example: `2026-04-14T12:00:00Z` |
| `json` | JSON object or JSON-serializable value | Returned as JSON in record responses |
| `uuid` | UUID string | Must be a valid UUID |

Unique constraints are expressed by appending `_u` to the type string.

Supported unique forms:

- `string_u`
- `int_u`
- `float_u`
- `uuid_u`

Unsupported unique forms such as `bool_u`, `time_u`, and `json_u` are rejected.

### Reserved Fields

You cannot define or write these fields yourself:

- `id`
- `created_at`
- `updated_at`

Baselix manages them automatically.

## Record Shape

Record responses are flat JSON objects.

Example:

```json
{
  "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "created_at": "2026-04-14T12:00:00Z",
  "updated_at": "2026-04-14T12:00:00Z",
  "email": "ada@example.com",
  "age": 34,
  "profile": {
    "plan": "pro"
  }
}
```

When a table has a field but a particular record has no stored value for it, Baselix returns that field as `null` on record fetches where the table schema is loaded.

## Creating Schemas

You have two ways to create a table:

### Explicit schema creation

Use `POST /api/v1/tables/{name}/schema` when you need exact types or unique constraints.

### Auto-created tables from record writes

If you `POST` records to a table that does not exist yet, Baselix infers a schema and creates the table automatically.

Auto-inference rules:

- Baselix merges fields from every record in the submitted batch.
- Strings that parse as UUIDs infer as `uuid`.
- Strings that parse as RFC3339 timestamps infer as `time`.
- JSON numbers infer as `float`.
- Boolean values infer as `bool`.
- JSON objects infer as `json`.
- List values are rejected.
- Unique constraints are not inferred.
- Mixed value types across the same field can lead to an unexpected schema.

If you care about integer semantics, unique fields, or predictable type control, create the schema explicitly first.

## Filtering

Filter syntax:

```text
filter=field:operator:value[,field:operator:value]
```

Multiple filter expressions are combined with AND.

### Operators

| Operator | Meaning | Works today | Notes |
| --- | --- | --- | --- |
| `eq` | Equality | Yes | Works on meta fields and custom fields |
| `neq` | Not equal | Yes | `nq` is also accepted |
| `nq` | Not equal | Yes | Short alias for `neq` |
| `gt` | Greater than | Yes | Requires type-compatible values |
| `gte` | Greater than or equal | Yes | Requires type-compatible values |
| `ge` | Greater than or equal | Yes | Alias for `gte` |
| `lt` | Less than | Yes | Requires type-compatible values |
| `lte` | Less than or equal | Yes | Requires type-compatible values |
| `le` | Less than or equal | Yes | Alias for `lte` |
| `contains` | Case-insensitive substring match | Yes | Only for `string` and `json` fields |
| `ne` | Not equal | No | Use `neq` or `nq` instead |

### Meta fields

You can filter these system fields directly:

- `id`
- `created_at`
- `updated_at`

### Null filtering

Current null behavior is intentionally documented as it exists today:

- `field:eq:null` works for user-defined fields and returns records where the field has no stored value.
- Null filters are not supported on `id`, `created_at`, or `updated_at`.
- `field:neq:null` and `field:nq:null` return records where the field exists.

### Type coercion during filtering

Filter values are coerced based on the field type:

- `int` expects a valid integer string.
- `float` expects a valid float string.
- `bool` expects `true` or `false`.
- `time` expects RFC3339.
- `uuid` expects a valid UUID string.

If coercion fails, the current runtime may return `500` instead of `400`.

## Sorting

Sort syntax:

```text
sort=field:direction[,field:direction]
```

Directions:

- `asc`
- `desc`

Sorting behavior:

- Multi-sort is stable.
- Invalid directions silently fall back to `asc`.
- `id`, `created_at`, and `updated_at` can always be used.
- Custom sort fields must exist in at least one returned record, otherwise sorting fails.
- Record-list responses are currently capped at `1000` items, so sorting happens within the returned result set.

## Limits and Plans

Baselix currently enforces three main categories of limits.

### Query limit

Record-list responses are currently capped at `1000` items for:

- `GET /api/v1/tables`
- `GET /api/v1/tables/{name}`

There is no cursor or offset pagination in the current API.

### Per-request record creation limit

A single `POST /api/v1/tables/{name}` request can currently include up to `1000` records.

### Plan limits

Plans are loaded from `plans.json`:

| Plan | Projects | Records per project |
| --- | --- | --- |
| `free` | 2 | 100 |
| `builder` | 5 | 1000 |
| `pro` | 20 | 100000 |

Important details:

- Project-count limits are enforced when creating projects from the dashboard.
- Record-count limits are enforced during record creation.
- Record updates do not perform the same plan-limit check.
- Exceeding the record-create limit returns `402 Payment Required`.

### Current mismatch in the limit error text

The create-record error message says Baselix is adding "values in request" to total project records. The current code actually counts existing project records plus submitted records. If you are building client messaging around that response, treat the message text as approximate.

## Error Model

Most API errors use this shape:

```json
{
  "error": 400,
  "message": "human-readable detail"
}
```

Examples:

- `401`: missing or invalid API key
- `402`: plan limit exceeded during record creation
- `404`: table or record not found
- `500`: internal error or a validation path that currently bubbles up too low in the stack

### Special-case error shape during auto-created table writes

If `POST /api/v1/tables/{name}` creates a table and then hits a unique-value error while inserting records, the response shape changes slightly:

```json
{
  "message": "Table 'customers' created with inferred schema.",
  "error": 400,
  "details": "duplicate value: 'ada@example.com' for unique field: email"
}
```

## Destructive Schema Operations

Schema updates and table deletion require:

```json
{
  "allow_destructive": true
}
```

### What `PUT /tables/{name}/schema` can destroy

- Fields removed from the payload are deleted.
- All values for deleted fields are deleted.
- If a field changes type, existing values for that field are deleted before the type is updated.
- Newly added fields appear as `null` on existing records until values are written.

### What `DELETE /tables/{name}/schema` destroys

- The table
- Its fields
- Its records
- All stored values for those records

## Current Runtime Caveats

These are the most important current API caveats to know before shipping a client:

- `PATCH /api/v1/tables/{name}` accepts a single JSON object or an array because it reuses the same body parser as record creation. Arrays are still the safest documented shape.
- Use `neq` or `nq` for not-equal filters. `ne` is not a supported operator.
- `field:eq:null` works for user-defined fields. `field:neq:null` and `field:nq:null` return records where the field exists.
- Invalid filter coercions such as a bad RFC3339 timestamp or `contains` on a `bool` field currently surface as `500`.
- Invalid schema type strings on table create/update currently surface as `500` after request parsing rather than a clean `400`.
- Updating a record with a field that is not in the table schema currently surfaces as `500`.
- There are no JSON-path query operators.
- Record-list responses are currently capped at `1000` items.

For the exact route contract, see [API Reference](./api.md).