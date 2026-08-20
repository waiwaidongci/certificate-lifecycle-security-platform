# Issuer list returns partial state

## Bug

The issuer list query treats a database row that cannot be scanned as a successful zero-value issuer. The response therefore contains a blank issuer while reporting the original row count as a successful total.

## Trigger

Create an `issuers` table without the production `NOT NULL` constraints, insert one valid row and one row with a NULL issuer field, and request the first issuer page through the list repository or service.

## Observed error

The list call returns two items and `total=2` with no error. The second item has an empty ID, name, provider, timestamps, and version instead of reporting that the row is malformed.
