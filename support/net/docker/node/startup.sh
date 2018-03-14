#!/bin/bash -e

FLAGS=${FLAGS:-"-sim_stdin=false"}

# replace the config values with the ones passed from env args
sed -i "/IdentityChainID/c\IdentityChainID = ${ID_CHAIN}" /root/.factom/m2/factomd.conf
sed -i "/LocalServerPublicKey/c\LocalServerPublicKey = ${PUB_KEY}" /root/.factom/m2/factomd.conf
sed -i "/LocalServerPrivKey/c\LocalServerPrivKey = ${PRIV_KEY}" /root/.factom/m2/factomd.conf

/go/bin/factomd $FLAGS
