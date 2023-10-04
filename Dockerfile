FROM golang:1.20-buster AS builder

# hadolint ignore=DL3008
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
  upx-ucl

WORKDIR /build

COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 go build \
      -ldflags='-w -s -extldflags "-static"' \
      -o ./bin/ctoc cmd/ctoc/main.go \
 && upx-ucl --best --ultra-brute ./bin/ctoc

FROM scratch
COPY --from=builder /build/bin/ctoc /bin/
WORKDIR /workdir
ENTRYPOINT ["/bin/ctoc"]
