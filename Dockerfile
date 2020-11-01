FROM golang:1.15-alpine3.12 AS build

WORKDIR /opt/app

ENV TMPDIR /tmp

RUN apk add make curl \
# remove spectral installation from Dockerfile
# when https://github.com/stoplightio/spectral/issues/1374 is fixed
 && apk add nodejs npm \
 && npm install -g @stoplight/spectral@5.6.0 \
# protoc
 && apk add protobuf-dev \
# buf
 && curl -sSL "https://github.com/bufbuild/buf/releases/download/v0.29.0/buf-$(uname -s)-$(uname -m)" -o /usr/bin/buf \
 && chmod +x /usr/bin/buf \
# yq
 && apk add yq --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community \
# done
 && echo 'done'

COPY go.* ./

RUN go mod download \
 && echo 'done'

COPY api/.spectral.yaml ./api/.spectral.yaml
COPY api/api.proto ./api/api.proto
COPY api/buf.yaml ./api/buf.yaml
COPY api/health.yaml ./api/health.yaml
COPY configs/ ./configs/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY main.go ./
COPY Makefile ./

RUN CGO_ENABLED=0 make build \
 && echo 'done'

FROM alpine:3.12 AS run

WORKDIR /opt/app

COPY --from=build /opt/app/bin/app ./

ENTRYPOINT ["./app"]
