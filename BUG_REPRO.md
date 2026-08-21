# Webhook rejection error chain

## Reproduction

Run the targeted test from the repository root:

```text
go test ./internal/notification/adapter/webhook -run '^TestSendPreservesWebhookRejectionChain$' -count=1
```

On the bug branch the test fails because a non-2xx webhook response is returned as a plain formatted error. `errors.As` cannot recover the HTTP delivery status and `errors.Is` cannot identify `ErrWebhookRejected`. The corrected implementation wraps `DeliveryError` with `%w`; its `Unwrap` method exposes the sentinel rejection cause.
