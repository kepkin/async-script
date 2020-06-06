FROM golang:1.12.6-stretch as build-env

RUN apt-get update && apt-get install time
# Install golangci-lint
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest && \
    golangci-lint --version

RUN mkdir -p /opt/billing
WORKDIR /opt/billing
COPY . .
RUN go build

FROM debian:stretch-slim

WORKDIR /opt/billing
COPY --from=build-env /opt/billing/async-script /usr/local/bin/async-script

ENTRYPOINT ["async-script"]
