# Reproduction

## Bug

When policy route dependencies are absent, registering a policy route can leave a handler that panics on the first request. A missing policy row scanner can also panic instead of returning a controlled application error.

## Trigger

Run the focused policy HTTP and SQLite scanner regression tests from the repository root. The test covers a nil handler, a handler without a service, a nil scanner interface, and a typed-nil scanner.

## Error

The HTTP case terminates with `panic: runtime error: invalid memory address or nil pointer dereference` from the policy handler. The scanner case reports an uncontrolled error after `nil policy row was scanned` panics.
