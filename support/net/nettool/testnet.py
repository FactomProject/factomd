"""
Manages the network of factomd nodes.
"""
from nettool import services


class Testnet(object):
    """
    Represents a factomd testnet running a set of factomd node and a seeds
    server.
    """
    def __init__(self, docker, config, network):
        self.network = network
        self.nodes = [services.Factomd(docker, cfg) for cfg in config]
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
