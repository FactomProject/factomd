"""
Manages the network of factomd nodes.
"""
from nettool import services
from nettool.docker_client import Image
from nettool.identities import IdentityPool


class Testnet(object):
    """
    Represents a factomd testnet running a set of factomd node and a seeds
    server.
    """
    def __init__(self, docker, config, flags, network):
        self.network = network

        self.base_factomd_image = Image(
            docker,
            tag="nettool_factomd_base",
            path="../../"
        )

        self.identity_pool = IdentityPool()
        self.nodes = []

        for cfg in config:
            identity = self.identity_pool.assign_next(cfg.name)
            node = services.Factomd(docker, cfg, identity, cfg.flags or flags)
            self.nodes.append(node)

        self.seeds = services.SeedServer(docker, self.nodes)

        for node in self.nodes:
            self.network.add(node)

        self.network.add(self.seeds)

    def print_info(self):
        """
        Prints the current status of the testnet.
        """
        self.seeds.print_info()

        for node in self.nodes:
            node.print_info()

    def up(self, build=False):
        """
        Brings the testnet up.
        """
        self.base_factomd_image.build(rebuild=build)

        if build:
            services.SeedServer.rebuild_image()
            services.Factomd.rebuild_image()

        self.seeds.generate_seeds_file()
        self.seeds.up(restart=build)

        for node in self.nodes:
            node.up(restart=build)

    def down(self, destroy=False):
        """
        Stops the testnet.
        """
        for node in self.nodes:
            node.down(destroy=destroy)

        self.seeds.down(destroy=destroy)

        if destroy:
            services.Factomd.destroy_image()
            services.SeedServer.destroy_image()
            self.base_factomd_image.destroy()
