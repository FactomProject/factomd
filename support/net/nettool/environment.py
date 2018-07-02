"""
Library for manipulating an environment hosting a network of factomd nodes.
"""
from nettool import log, docker_client, testnet, rules, network


class Environment(object):
    """
    Represents an environment hosting a factomd testnet along with supporting
    services.
    """
    def __init__(self, config):
        docker = docker_client.create()
        self.network = network.Network(docker)
        self.testnet = testnet.Testnet(docker, config, self.network)
        self.rules = rules.Rules(
            docker,
            config.network,
            self.network,
            self.testnet
        )

    def print_info(self):
        """
        Prints the current status of the environment.
        """
        log.section("Info")

        self.testnet.print_info()
        self.rules.print_info()

    def up(self, build_mode=False):
        """
        Ensures that all necessary components in the environment are up and
        running. If build_mode pareamter is set to True, forces a rebuild of
        existing images.
        """
        log.section("Starting the environment")

        self.network.up(build=build_mode)
        self.rules.up(build=build_mode)
        self.testnet.up(build=build_mode)

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

        self.testnet.down(destroy=destroy_mode)
        self.rules.down(destroy=destroy_mode)
        self.network.down(destroy=destroy_mode)
