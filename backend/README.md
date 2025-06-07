# OTLP demo backend

This is a showcase on how to setup a secured Open Telemetry Collector backend using

-   Loki
-   Tempo
-   Prometheus
-   Grafana

This setup should use mTLS where possible. Therefore the certificates created at `certs/docker` are used. To keep the demo simple we use one server cert with the SAN for **loki**,**tempo** and **prometheus** and one common client cert to verify the client.

## Requirements

We use self signed certificates here. Use [this script](../createCerts.sh) to create all required certs.

## Observations

While everything is working there is some unwanted info message in the grafana logs with this setup. It may not directly be related to the mTLS setup.

```console
{
  "logger": "grafana-apiserver",
  "t": "2025-06-07T20:22:48.635503532Z",
  "level": "info",
  "msg": "[core] [Channel #2 SubChannel #3]grpc: addrConn.createTransport failed to connect to {Addr: \\\"tempo:3200\\\", ServerName: \\\"tempo:3200\\\", BalancerAttributes: {\\\"<%!p(pickfirstleaf.managedByPickfirstKeyType={})>\\\": \\\"<%!p(bool=true)>\\\" }}. Err: connection error: desc = \\\"error reading server preface: EOF\\\""
}
```

In a non-TLS setup this message is observed when `stream_over_http_enabled` evaluates to `false`. Setting it to `true` prevents above info message. This is independent of the streaming settings in the grafana datasource.
