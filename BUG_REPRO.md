# Bug Reproduction

## Bug

Terminal lifecycle states are classified inconsistently, and completed rotation work can reopen as pending.

## Trigger

Run `go test ./internal/rotation/domain -run '^TestLifecycleTerminalStateContract$' -count=1` in `env/`.

## Error

The test reports that expired certificates and failed reminders are rejected, skipped rotations are not terminal, and terminal rotations can transition back to pending.
