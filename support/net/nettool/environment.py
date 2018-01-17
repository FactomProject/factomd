"""
Library for manipulating an environment hosting a network of factomd nodes.
"""
from nettool import log, network, services


class Environment(object):
    """
    Represents an environment hosting a  network of factomd nodes along with
    some supporting services.
    """
    def __init__(self, config, docker):
        self.config = config
        self.docker = docker
        self.network = network.Network(self)
        self.gateway = services.Gateway(self)
        self.seeds = services.SeedServer(self)
        self.nodes = []
        self._populate_nodes()

    def print_info(self):
        """
        Prints the current status of the environment.
        """
        log.section("Info")

        for container in self._containers:
            container.print_info(self.docker)

    def up(self, build_mode=False):
        """
        Ensures that all necessary components in the environment are up and
        running. If build_mode pareamter is set to True, forces a rebuild of
        existing images.
        """
        log.section("Starting the environment")
        if build_mode:
            for image in self._images:
                image.build(self.docker, rebuild=True)

        self.network.up(self.docker)

        for container in self._containers:
            container.up(self.docker, restart=build_mode)

    def down(self, destroy_mode=False):
        """
        Ensures that all environment components are stopped. If destroy_mode
        parameter is set to True, also removes all previously created
        containers and images.
        """
        if destroy_mode:
            log.section("Destroying the environment")
        else:
            log.section("Stopping the environment")

        for container in self._containers:
            container.down(self.docker, destroy=destroy_mode)

        self.network.down(self.docker, destroy=destroy_mode)

        if destroy_mode:
            for image in self._images:
                image.destroy(self.docker)

    def _populate_nodes(self):
        for node_cfg in self.config.nodes:
            node = services.Factomd(node_cfg, self)
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
