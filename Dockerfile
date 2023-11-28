FROM alpine:3.18.4
RUN apk upgrade --update --no-cache
RUN apk add --update --no-cache ca-certificates haproxy
COPY hcloud-talos-controlplane-gateway /bin/hcloud-talos-controlplane-gateway
ENTRYPOINT ["/bin/hcloud-talos-controlplane-gateway", "start"]
