# TLS DEMO

Certificates should have been created via the [createCerts.sh](./createCerts.sh) skript.

Default settings:

```console
insecure: true
tls-skip-verify: false
log-level: debug
addr: localhost:8080
```

## No TLS

### Server

```console
go run main.go web webserver
```

### Client

```console
go run main.go web tlsclient --url http://localhost:8080/hello
```

## Use server certifcate

### Server

```console
go run main.go web webserver --insecure=false --tls-key certs/server.key --tls-cert certs/server.crt
```

### Client

We now have to use TLS. From here on we need to address the server via **https**.
The server uses a self signed cert. We can either ignore the tls verification or provide the rootCA that was used to create the server cert.

```console
go run main.go web tlsclient --url https://localhost:8080/hello --insecure=false --tls-skip-verify
go run main.go web tlsclient --url https://localhost:8080/hello --insecure=false --tls-ca certs/rootCA.pem
```

## Use mTLS (simple)

The client has to provide a cert that the server can verify. The server decides by configuration which method should be used. This is done by the parameter `--tls-client-auth`.
A strict variant would be the value `require-and-verify`. At this point we have to provide a CA on the server side to verify the client cert.

In the current setup we assume that client and server certs share the same CA `certs/rootCA.pem`

### Server

```console
go run main.go web webserver --insecure=false --tls-key certs/server.key --tls-cert certs/server.crt --tls-client-ca certs/rootCA.pem --tls-client-auth require-and-verify
```

### Client

Here we provide the client cert via `--tls-key` and `--tls-cert`. The parameter `--tls-ca` is used to verify the server cert.

```console
go run main.go web tlsclient --url https://localhost:8080/hello --insecure=false --tls-ca certs/rootCA.pem --tls-key certs/client.key --tls-cert certs/client.crt
```

## Use mTLS (advanced)

In this example we use a separate CA for the client certs. These are store in [certs/special](./certs/special/)

### Server

The server needs to know the CA for the client certs via the `--tls-ca` parameter

```console
go run main.go web webserver --insecure=false --tls-key certs/server.key --tls-cert certs/server.crt --tls-client-ca certs/special/clientRootCA.pem --tls-client-auth require-and-verify
```

### Client

Here we provide the client cert via `--tls-key` and `--tls-cert`. Note, here we use the files in **certs/special**
The parameter `--tls-ca` is used to verify the server cert.

```console
go run main.go web tlsclient --url https://localhost:8080/hello --insecure=false --tls-ca certs/rootCA.pem --tls-key certs/special/client.key --tls-cert certs/special/client.crt
```
