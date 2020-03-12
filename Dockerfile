FROM golang:1.13

# Get git
RUN apt-get update \
    && apt-get -y install curl git \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Where factomd sources will live
WORKDIR /opt/FactomProject/factomd

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
