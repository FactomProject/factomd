"""
Library for manipulating an environment hosting a network of factomd nodes.
"""
from nettool import log, container, network, services


class Environment(object):
    """
    Represents an environment hosting a network of factomd nodes along with
    supporting services.
    """
    def __init__(self, config, docker):
        container.Container.env = self
        network.Network.env = self

        self.config = config
        self.docker = docker
        self.gateway = services.Gateway()
        self.nodes = [services.Factomd(node) for node in self.config.nodes]
        self.seeds = services.SeedServer(self.nodes)
        self.network = network.Network(config.network, self._containers)

    def print_info(self):
        """
        Prints the current status of the environment.
        """
        log.section("Info")

        for container in self._containers:
            container.print_info()

        self.network.print_info()

    def up(self, build_mode=False):
        """
        Ensures that all necessary components in the environment are up and
        running. If build_mode pareamter is set to True, forces a rebuild of
        existing images.
        """
        log.section("Starting the environment")

        if build_mode:
            for image in self._images:
                image.build(rebuild=True)

        self.network.up()
        self.seeds.generate_seeds_file()

        for container in self._containers:
            container.up(restart=build_mode)

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
            container.down(destroy=destroy_mode)

        self.network.down(destroy=destroy_mode)

        if destroy_mode:
            for image in self._images:
                image.destroy()

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
