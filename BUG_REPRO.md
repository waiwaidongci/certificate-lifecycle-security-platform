# Bug reproduction

## Bug

The issuer list filter silently treats an invalid `enabled` value as false instead of returning a validation error.

## Trigger

Call the issuer list query path with `enabled=sometimes` and observe the repository filter handling.

## Error

The request is accepted and the query is built with `enabled = 0`; the caller does not receive an `INVALID_ARGUMENT` error.
