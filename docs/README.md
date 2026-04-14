# Baselix API Documentation

This folder documents the current Baselix HTTP API. The API is project-scoped, JSON-based, and centered around tables and records.

## Start Here

- [Quickstart](./quickstart.md)
- [Concepts and Caveats](./concepts.md)
- [API Reference](./api.md)
- [OpenAPI 3.0 Spec](./openapi.yaml)
- [Postman Collection](./Baselix%20API.postman_collection.json)

## Scope

This documentation covers:

- `GET /ping`
- All authenticated routes under `/api/v1`
- How project API keys are created and rotated from the dashboard

This documentation does not cover:

- Clerk webhook handling
- HTML dashboard routes as a public API surface
- Frontend templates

## Base URLs

Baselix is a hosted SaaS service. Use:

- App root: `https://baselix.dev`
- API root: `https://baselix.dev/api/v1`

## Authentication

All `/api/v1` routes require a project API key:

```http
Authorization: Bearer <api_key>
```

API keys are created and rotated from the signed-in dashboard, not from the REST API. Treat the displayed key as a secret and keep a copy when you create or rotate it.

## What Baselix Exposes

- Table schema management
- Record CRUD
- Auto-created tables from record payloads
- Filtering and multi-sort
- Per-project isolation behind the API key
- Plan and request-size limits

## Important Runtime Notes

- There is no cursor or offset pagination in the current API.
- List responses are currently capped at `1000` items.
- Destructive schema updates and table deletion require `"allow_destructive": true`.
- Record responses are flat JSON objects.
- A few malformed inputs currently surface as `500` instead of `400`; see [Concepts and Caveats](./concepts.md) before building client-side validation around strict status-code assumptions.