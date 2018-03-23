"""
Manages the network of factomd nodes.
"""
import time

from nettool import services
from nettool.docker_client import Image
from nettool.identities import IdentityPool
from nettool import log


class Testnet(object):
    """
    Represents a factomd testnet running a set of factomd node and a seeds
    server.
    """
    def __init__(self, docker, config, network):
        self.network = network

        self.base_factomd_image = Image(
            docker,
            tag="nettool_factomd_base",
            path=config.factomd_path
        )

        self.identities = IdentityPool()
        self.nodes = list(self._create_nodes(docker, config))
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

        self.nodes[0].up(restart=build)
        self.nodes[0].load_identities(len(self.nodes) - 1)

        with log.step("WAITING"):
            time.sleep(120)

        for node in self.nodes[1:]:
            self.nodes[0].promote(node)

        for node in self.nodes[1:]:
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

    def _create_nodes(self, docker, config):
        # first node needs to get the bootstrap identity
        yield services.Factomd(
            docker,
            config.nodes[0],
            self.identities.bootstrap,
            config.nodes[0].flags or config.flags
        )

        # other nodes get identities from the pool
        for cfg in config.nodes[1:]:
            identity = self.identities.assign_next(cfg.name)
            yield services.Factomd(
                docker,
                cfg,
                identity,
                cfg.flags or config.flags)
