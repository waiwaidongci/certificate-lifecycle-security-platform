# Bug reproduction

## Bug

Concurrent requests for the same service, template, target type, and target create independent distribution records and send the same configuration twice.

## Trigger

Start one distribution request and hold its adapter call open. Submit an identical distribution request before the first adapter call completes.

## Observed error

The duplicate request succeeds instead of returning a conflict. The adapter call count is `2`, and two distribution records are persisted for the same in-progress delivery.
