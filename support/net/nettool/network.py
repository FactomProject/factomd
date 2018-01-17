"""
Module for manipulating the docker network that allows connectivity between
nodes.
"""
import docker as docker_lib

from nettool import log


NETWORK_NAME = "nettool"


class Network(object):

    def __init__(self, env):
        self.env = env
        self.name = NETWORK_NAME
        self.docker_network = None

    def is_up(self, docker):
        self._refresh_network_status(docker)
        return self.docker_network is not None

    def up(self, docker):
        if self.is_up(docker):
            return

        with log.step("Creating network"):
            ipam_pool = docker_lib.types.IPAMPool(
                subnet="10.12.0.0/16",
                gateway="10.12.0.254",
                iprange="10.12.1.0/24"
            )
            ipam_config = docker_lib.types.IPAMConfig(
                pool_configs=[ipam_pool]
            )
            self.docker_network = docker.networks.create(
                self.name,
                driver='bridge',
                ipam=ipam_config
            )

    def down(self, docker, destroy=False):
        if destroy and self.is_up(docker):
            with log.step("Removing network"):
                self.docker_network.remove()
                self.docker_network = None

    def _refresh_network_status(self, docker):
        try:
            self.docker_network = docker.networks.get(self.name)
        except docker_lib.errors.NotFound:
            self.docker_network = None
