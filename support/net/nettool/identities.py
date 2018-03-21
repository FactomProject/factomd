"""
Management of custom network leader identities.
"""
from collections import deque, namedtuple
import yaml


IDENTITIES_FILE = "docker/identities.yml"


Identity = namedtuple("Identity", "chain, priv, pub")


BOOTSTRAP_IDENTITY = Identity(
    chain="38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9",
    priv="4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d",
    pub="cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a"
)


class IdentityPool(object):
    """
    A pool of preprepared identities that can be used to set up leaders in
    a custom network.
    """
    def __init__(self):
        self.identities = deque(_load_identities_from_file())
        self.assigned = {}
        self.bootstrap = BOOTSTRAP_IDENTITY

    def assign_next(self, node_name):
        """
        Assign an identity to a node identitified by its name.
        """
        if node_name in self.assigned.keys():
            raise Exception(
                f"{node_name} already has an identity assigned"
            )
        if not self.identities:
            raise Exception(
                f"No more identites to assign, add more in {IDENTITIES_FILE}"
            )

        identity = self.identities.popleft()
        self.assigned[node_name] = identity
        return identity

    def get_identity_for_node(self, node_name):
        """
        Returns an identity for the given node or None if the node wasn't
        assigned any identity yet.
        """
        return self.assigned.get(node_name)


def _load_identities_from_file():
    with open(IDENTITIES_FILE) as ident_file:
        for identity_cfg in yaml.load(ident_file):
            yield Identity(
                chain=identity_cfg["IdentityChainID"],
                priv=identity_cfg["ServerPrivKey"],
                pub=identity_cfg["ServerPublicKey"]
            )
