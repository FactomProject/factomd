"""
Library for manipulating a network of factomd nodes.
"""
from nettool import log, services


class Network(object):
    """
    Represents a network of factomd nodes along with some of the supporting
    services.
    """
    def __init__(self, config, docker):
        self.config = config
        self.docker = docker
        self.gateway = services.Gateway()
        self.seeds = services.SeedServer()
        self.nodes = []
        self._populate_nodes()

    def print_status(self):
        """
        Prints the current status of the network.
        """
        log.section("Network status")

        for container in self._containers:
            container.print_status(self.docker)

    def up(self, build_mode=False):
        """
        Ensures that all necessary components in the network are up and
        running. If build_mode pareamter is set to True, forces a rebuild of
        existing images.
        """
        log.section("Starting the network")
        if build_mode:
            for image in self._images:
                image.build(self.docker, rebuild=True)

        for container in self._containers:
            container.up(self.docker, restart=build_mode)

    def down(self, destroy_mode=False):
        """
        Ensures that all network components are stopped. If destroy_mode
        parameter is set to True, also removes all previously created
        containers and images.
        """
        if destroy_mode:
            log.section("Destroying all artifacts")
        else:
            log.section("Stopping the network")

        for container in self._containers:
            container.down(self.docker, destroy=destroy_mode)

        if destroy_mode:
            for image in self._images:
                image.destroy(self.docker)

    def _populate_nodes(self):
        for node_cfg in self.config.nodes:
            node = services.Factomd(node_cfg)
            self.nodes.append(node)
            if node_cfg.seed:
                self.seeds.add(node)

    @property
    def _containers(self):
        yield self.gateway
        yield self.seeds

        for node in self.nodes:
            yield node

    @property
    def _images(self):
        return [
            services.Gateway,
            services.SeedServer,
            services.Factomd
        ]
