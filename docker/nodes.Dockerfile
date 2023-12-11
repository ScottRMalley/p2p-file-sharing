FROM golang:1.20-buster AS builder

ENV GO111MODULE "on"

WORKDIR $GOPATH/github.com/scottrmalley/p2p-file-challenge
ADD . ./

RUN GOOS=linux go build -a -o /node ./cmd/node

### Final Image
from debian:buster

LABEL Maintainer="Scott R. Malley <scott.r.malley@gmail.com>"
LABEL Name=""

COPY --from=builder /node /usr/bin/node

ENTRYPOINT ["/usr/bin/node"]
