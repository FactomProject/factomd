FROM nettool_factomd_base:latest

# compile addservermessage
RUN go get github.com/FactomProject/addservermessage

WORKDIR $GOPATH/src/github.com/FactomProject/addservermessage

RUN glide install -v

RUN go install


# restore factomd workdir
WORKDIR $GOPATH/src/github.com/FactomProject/factomd
