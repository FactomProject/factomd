FROM nettool_factomd_base:latest

# compile addservermessage
RUN go get github.com/FactomProject/addservermessage

WORKDIR $GOPATH/src/github.com/FactomProject/addservermessage

RUN glide install -v

RUN go install


# overwrite the default config values
COPY factomd.conf /root/.factom/m2/factomd.conf

ARG ID_CHAIN
ARG PUB_KEY
ARG PRIV_KEY

RUN sed -i "/IdentityChainID/c\IdentityChainID = ${ID_CHAIN}" /root/.factom/m2/factomd.conf
RUN sed -i "/LocalServerPublicKey/c\LocalServerPublicKey = ${PUB_KEY}" /root/.factom/m2/factomd.conf
RUN sed -i "/LocalServerPrivKey/c\LocalServerPrivKey = ${PRIV_KEY}" /root/.factom/m2/factomd.conf

# restore factomd workdir
WORKDIR $GOPATH/src/github.com/FactomProject/factomd
