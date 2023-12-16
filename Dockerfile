# This file is used by goreleaser
FROM scratch
ENTRYPOINT ["/otlpdemo"]
COPY otlpdemo /
