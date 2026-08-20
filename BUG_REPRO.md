# Bug reproduction

## Bug

The metrics registry stores counters and histogram series in ordinary maps. First-use registration and the HTTP metrics handler enumerate those maps concurrently, so the process can report a data race or terminate with `concurrent map read and map write`.

## Trigger

Start several goroutines that create distinct counter and histogram names while another goroutine repeatedly serves the metrics endpoint. The test uses eight writers and one reader with a synchronized start so the first registration and exposition overlap.

## Observed error

The race-enabled test reports conflicting reads and writes in `Metrics.ensure` and `Metrics.Handler`, and the baseline can terminate with `fatal error: concurrent map read and map write`. The final exposition is not reliable because registration can race with enumeration.
