FROM golang:1.22 as build
COPY . .
RUN CGO_ENABLED=0 go build

FROM debian:bookworm-slim


COPY --from=build /go/webshell-proxy /usr/bin/webshell-proxy

RUN useradd -ms /bin/bash webshell-proxy

ENTRYPOINT [ "/usr/bin/webshell-proxy" ]
