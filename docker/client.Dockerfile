FROM golang:1.20-buster AS builder

RUN apt update && apt install -y git build-essential gcc g++

ENV GO111MODULE "on"

WORKDIR $GOPATH/github.com/scottrmalley/p2p-file-challenge
ADD . ./

RUN GOOS=linux go build -a -o /client ./cmd/client

### Final Image
from debian:buster

LABEL Maintainer="Scott R. Malley <scott.r.malley@gmail.com>"
LABEL Name=""

COPY --from=builder /client /usr/bin/client

ENTRYPOINT ["/usr/bin/client"]
