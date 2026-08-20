# Issuer SAN slice aliasing

The issuer provider normalizes SAN values in place. Because the request crosses HTTP, service, domain, and provider boundaries without a consistent ownership rule, the caller's slice can change unexpectedly.

The focused tests in `collection.json` fail on the bug baseline and pass after each boundary takes an independent snapshot.
