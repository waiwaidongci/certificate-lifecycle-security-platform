# Bug Reproduction

## Bug

Several certificate platform creation endpoints classify malformed JSON as an internal server error.

## Trigger

Run `go test ./internal/certificate/adapter/http -run '^TestMalformedCreateRequestsKeepInvalidArgumentResponse$' -count=1` in `env/`.

## Error

Each subtest receives HTTP 500 with `{"code":"INTERNAL","message":"internal error"}` instead of HTTP 400 and `INVALID_ARGUMENT`.
