#!/bin/bash

# Sample script to create self-signed certificates for server and client
# The script creates a rootCA, server and client certificates for localhost

# ensure the output directory exists
mkdir -p certs/special
mkdir -p certs/docker

# $PASS contains the password in the form "pass:password"


# create rootCA private key
openssl genpkey -algorithm RSA -out certs/rootCA.key -aes256 -pass $PASS

# self signed rootCA
openssl req -x509 -new -nodes -key certs/rootCA.key -sha256 -days 3650 -subj "/CN=otlpdemo root CA/O=otlpdemo" -passin $PASS -out certs/rootCA.pem

## server certificate

# create server private key
openssl genpkey -algorithm RSA -out certs/server.key
# create server certificate signing request
openssl req -new -key certs/server.key -out certs/server.csr -subj "/CN=somelocalhost/O=otlpdemo" \
 -addext 'subjectAltName = DNS:localhost, IP:127.0.0.1' \
 -addext "keyUsage = digitalSignature, keyEncipherment" \
 -addext "extendedKeyUsage = serverAuth"

# sign server certificate
openssl x509 -req -in certs/server.csr -CA certs/rootCA.pem -CAkey certs/rootCA.key -CAcreateserial -out certs/server.crt \
 -days 365 -sha256 \
 -copy_extensions copy \
 -passin $PASS

## client certificate

# create client private key
openssl genpkey -algorithm RSA -out certs/client.key
# create client certificate signing request
openssl req -new -key certs/client.key -out certs/client.csr -subj "/CN=democlient/O=otlpdemo" \
 -addext "keyUsage = digitalSignature, keyEncipherment" \
 -addext "extendedKeyUsage = clientAuth"
# sign server certificate
openssl x509 -req -in certs/client.csr -CA certs/rootCA.pem -CAkey certs/rootCA.key -CAcreateserial -out certs/client.crt \
 -days 365 -sha256 -copy_extensions copy  -passin $PASS


# special: client cert with own CA

# create clientRootCA private key
openssl genpkey -algorithm RSA -out certs/special/clientRootCA.key -aes256 -pass $PASS

# self signed clientRootCA
openssl req -x509 -new -nodes -key certs/special/clientRootCA.key -sha256 -days 3650 -subj "/CN=otlpdemo client root CA/O=otlpdemo clients" -passin $PASS -out certs/special/clientRootCA.pem

## client certificate

# create client private key
openssl genpkey -algorithm RSA -out certs/special/client.key
# create client certificate signing request
openssl req -new -key certs/special/client.key -out certs/special/client.csr -subj "/CN=special democlient/O=otlpdemo clients" \
 -addext "keyUsage = digitalSignature, keyEncipherment" \
 -addext "extendedKeyUsage = clientAuth"
# sign server certificate
openssl x509 -req -in certs/special/client.csr -CA certs/special/clientRootCA.pem -CAkey certs/special/clientRootCA.key -CAcreateserial -out certs/special/client.crt \
 -days 365 -sha256 -copy_extensions copy  -passin $PASS


# for mtls within docker compose setup we use dedicated certs

# create rootCA private key
openssl genpkey -algorithm RSA -out certs/docker/rootCA.key -aes256 -pass $PASS

# self signed rootCA
openssl req -x509 -new -nodes -key certs/docker/rootCA.key -sha256 -days 3650 -subj "/CN=otlpdemo docker root CA/O=otlpdemo docker" -passin $PASS -out certs/docker/rootCA.pem

## server certificate

# create server private key
openssl genpkey -algorithm RSA -out certs/docker/server.key
# create server certificate signing request
openssl req -new -key certs/docker/server.key -out certs/docker/server.csr -subj "/CN=common docker cert/O=otlpdemo" \
 -addext 'subjectAltName = DNS:prometheus, DNS:tempo, DNS:loki' \
 -addext "keyUsage = digitalSignature, keyEncipherment" \
 -addext "extendedKeyUsage = serverAuth"

# sign server certificate
openssl x509 -req -in certs/docker/server.csr -CA certs/docker/rootCA.pem -CAkey certs/docker/rootCA.key -CAcreateserial -out certs/docker/server.crt \
 -days 365 -sha256 \
 -copy_extensions copy \
 -passin $PASS

## client certificate

# create client private key
openssl genpkey -algorithm RSA -out certs/docker/client.key
# create client certificate signing request
openssl req -new -key certs/docker/client.key -out certs/docker/client.csr -subj "/CN=otlpclient/O=otlpdemo" \
 -addext "keyUsage = digitalSignature, keyEncipherment" \
 -addext "extendedKeyUsage = clientAuth"
# sign client certificate
openssl x509 -req -in certs/docker/client.csr -CA certs/docker/rootCA.pem -CAkey certs/docker/rootCA.key -CAcreateserial -out certs/docker/client.crt \
 -days 365 -sha256 -copy_extensions copy  -passin $PASS
