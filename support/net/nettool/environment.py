"""
Library for manipulating an environment hosting a network of factomd nodes.
"""
from nettool import log, services
import nettool.container
import nettool.iptables
import nettool.network


class Environment(object):
    """
    Represents an environment hosting a network of factomd nodes along with
    supporting services.
    """
    def __init__(self, config, docker):
        nettool.container.Container.env = self
        nettool.network.Network.env = self
        nettool.iptables.Iptables.env = self

        self.config = config
        self.docker = docker
        self.gateway = services.Gateway()
        self.nodes = [services.Factomd(node) for node in self.config.nodes]
        self.seeds = services.SeedServer(self.nodes)
        self.network = nettool.network.Network(self._containers)
        self.iptables = nettool.iptables.Iptables(
            self.gateway,
            self.config.network
        )

    def print_info(self):
        """
        Prints the current status of the environment.
        """
        log.section("Info")

        self.gateway.print_info()

        for container in self._containers:
            container.print_info()

        self.network.print_info()
        self.iptables.print_info()

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
        self.gateway.up()
        self.iptables.up()
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
        self.iptables.down()
        self.gateway.down(destroy=destroy_mode)

        if destroy_mode:
            for image in self._images:
                image.destroy()

    def ins_rule(self, source, target, action):
        """
        Insert a rule at the beginning of the chain.
        """
        self._ensure_gateway()
        self.iptables.ins_rule(source, target, action)

    def add_rule(self, source, target, action):
        """
        Append a rule at the end of the chain.
        """
        self._ensure_gateway()
        self.iptables.add_rule(source, target, action)

    def del_rule(self, source, target, action):
        """
        Delete a rule from the chain.
        """
        self._ensure_gateway()
        self.iptables.del_rule(source, target, action)

    def _ensure_gateway(self):
        if not self.gateway.is_running:
            log.fatal("Gateway container must be up to del rules")

    @property
    def _containers(self):
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
