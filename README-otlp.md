# Open telemetry

In a nutshell [Open Telemetry][otel] is a framework for observability. It handles tasks about creating, collecting and exporting telemetry data like metrics, traces and logs.

In this demo we use it to instrument the application. The data should be sent to an open telemetry collector. A sample configuration is available in the [backend](./backend/) via docker compose.

## Send data to Open telemetry collector

When enabling telemetry via `--enable-telemetry` the configuration and extra data is read from environment variables which are defined by OTLP.

See the [base config](./otlp-env.env) for a simple configuration without TLS and the [advanced config](./otlp-env-secure.env) for connections with mTLS.

### Configuration

A lot of configuration can be provided via environment variables. These start with the prefix `OTEL_`.
See [this link][otel-cfg-general] for settings about configuring what to collect.
See [this link][otel-cfg-exporter] for settings about configuring the exporter.

**Note:**
As of now (2025-06-01) the log export via gRPC is not directly supported via `OTEL_EXPORT_` vars when using TLS. This has been added in this application as a workaround in this [ticket](https://github.com/mpapenbr/otlpdemo/issues/69)

---

[otel]: https://opentelemetry.io/docs/what-is-opentelemetry/
[otel-cfg-general]: https://opentelemetry.io/docs/languages/sdk-configuration/general/
[otel-cfg-exporter]: https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/
