# Bug Reproduction

## Bug

Concurrent notification workers can claim the same pending reminder because the read, send, and status transition are not coordinated. The same reminder is delivered more than once.

## Trigger

Run the concurrent worker verification test from the project root with the race detector. The test starts two workers against the same pending reminder and observes the delivery count.

## Error

`send_pending_test.go:79: same pending reminder was delivered 2 times`
