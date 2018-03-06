"""
Management of custom network leader identities.
"""
from collections import namedtuple
import yaml


IDENTITIES_FILE = "docker/identities.yml"


Identity = namedtuple("Identity", "chain_id, priv_key, pub_key")


class IdentityPool(object):
    """
    A pool of preprepared identities that can be used to set up leaders in
    a custom network.
    """
    env = None

    def __init__(self):
        self.identities = list(_load_identities_from_file())
        self.assigned = {}

    def assign_next_identity(self, node_name):
        """
        Assign an identity to a node identitified by its name.
        """
        if node_name in self.assigned.keys():
            raise Exception(
                "{} already has an identity assigned" % node_name
            )
        if not self.identities:
            raise Exception(
                "No more identites to assign, add more in {}" % IDENTITIES_FILE
            )

        identity = self.identities.pop()
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
                chain_id=identity_cfg["IdentityChainID"],
                priv_key=identity_cfg["ServerPrivKey"],
                pub_key=identity_cfg["ServerPublicKey"]
            )
