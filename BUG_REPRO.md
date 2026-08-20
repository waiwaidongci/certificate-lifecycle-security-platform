# Reproduction

## Bug

A cancelled certificate issuance can continue through detached database contexts and persist certificate, service-binding, or event-log state.

## Trigger

Run the targeted certificate cancellation test in the bug branch.

## Error

`Issue error = <nil>, want context cancellation`
