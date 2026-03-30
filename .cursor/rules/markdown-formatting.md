---
inclusion: fileMatch
fileMatchPattern: "**/*.md"
---

# Markdown Formatting Rules

## Forbidden: Pipe tables (`|`)

NEVER use pipe-format tables in `.md` files:

```
| Column | Column |
|---|---|
| value | value |
```

This format breaks and becomes impossible to read in many contexts.

## Alternative: Descriptive lists

Use simple key-value lists:

```
- **provider.go** — Initializes MeterProvider with resource attributes
- **metrics.go** — Generic implementation of Counter, Histogram, Gauge
```

Or descriptive blocks when more detail is needed:

```
`provider.go`
  Initializes MeterProvider with resource attributes (service name, version, environment).
  Status: Ready

`metrics.go`
  Generic implementation of Counter, Histogram, Gauge via observability.Metrics interface.
  Status: Ready
```
