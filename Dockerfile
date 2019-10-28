FROM golang:1.12

# Get git
RUN apt-get update \
    && apt-get -y install curl git \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Get glide
RUN go get github.com/Masterminds/glide

# Where factomd sources will live
WORKDIR $GOPATH/src/github.com/FactomProject/factomd

# Get the dependencies
COPY glide.yaml glide.lock ./

# Install dependencies
RUN glide install -v

# Get goveralls for testing/coverage
RUN go get github.com/mattn/goveralls

# Populate the rest of the source
COPY . .

ARG GOOS=linux

# Build and install factomd
RUN ./build.sh

# Setup the cache directory
RUN mkdir -p /root/.factom/m2
COPY factomd.conf /root/.factom/m2/factomd.conf

ENTRYPOINT ["/go/bin/factomd","-sim_stdin=false"]

EXPOSE 8088 8090 8108 8109 8110
