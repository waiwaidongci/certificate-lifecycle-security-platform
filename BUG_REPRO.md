# Notification cancellation regression

## Reproduction

Run the two targeted tests from the repository root:

```text
go test ./internal/notification/infrastructure/sqlite -run '^TestPendingReminderOperationsHonorCancellation$' -count=1
go test ./internal/notification/adapter/log -run '^TestLogSendHonorsCancellation$' -count=1
```

On the bug branch, both tests fail because a cancelled context still allows the pending reminder operation or log delivery to complete. The SQLite test reports `ListPending error = <nil>, want context cancellation`; the adapter test reports `Send error = <nil>, want context cancellation`. The reminder remains pending in the database check, but the cancelled operation incorrectly proceeds instead of returning `context.Canceled`.

The corrected implementation checks the caller context before and during pending reminder database work and before notification delivery, while passing the original context through to the underlying operation.
