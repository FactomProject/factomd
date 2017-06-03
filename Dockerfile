FROM golang:1.8.3-alpine

# Get git
RUN apk add --no-cache curl git

# Get glide
RUN go get github.com/Masterminds/glide

# Where factomd sources will live
WORKDIR $GOPATH/src/github.com/FactomProject/factomd

# Populate the source
COPY . .

# Install dependencies
RUN glide install -v

ARG GOOS=linux

# Build and install factomd
RUN go install

# Setup the cache directory
RUN mkdir -p $HOME/.factom/m2
COPY factomd.conf $HOME/.factom/m2/factomd.conf

ENTRYPOINT ["/go/bin/factomd"]

EXPOSE 8088 8090 8108 8109 8110
