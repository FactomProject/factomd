"""
Module for manipulating the docker network that allows connectivity between
nodes.
"""
from ipaddress import ip_address, ip_network
import docker as docker_lib

from nettool import log

# docker network name
NETWORK_NAME = "nettool"

# range of addresses that will be used to assign IPs to containers
NETWORK_SUBNET = "10.12.0.0/24"

# the IP of the network gateway for the subnet above
NETWORK_GATEWAY = "10.12.0.254"

# the range of addresses that docker internal DHCP can use, should be different
# from the subnet above
NETWORK_IPRANGE = "10.12.1.0/24"


class Network(object):
    """
    A docker network that takes care of static IP assignment and allows
    connectivity between containers.
    """
    env = None

    def __init__(self, docker):
        self.name = NETWORK_NAME
        self.docker = docker
        self.docker_network = None
        self.ip_pool = IPPool()

    @property
    def address(self):
        """
        Get the address of the network.
        """
        return ip_network(NETWORK_SUBNET)

    def add(self, service):
        """
        Adds a container to the containers managed by the network.
        """
        ip = self.ip_pool.add(service.container)
        service.container.assigned_ip = ip
        service.container.network = self

    def is_up(self):
        """
        Checks if the docker network is up.
        """
        self._refresh_network_status()
        return self.docker_network is not None

    def up(self, build=False):
        """
        Ensures that the network is running.
        """
        if build:
            self.down(destroy=True)

        if self.is_up():
            return

        with log.step("Creating network"):
            ipam_pool = docker_lib.types.IPAMPool(
                subnet=str(self.ip_pool.subnet),
                gateway=str(self.ip_pool.gateway),
                iprange=str(self.ip_pool.iprange)
            )
            ipam_config = docker_lib.types.IPAMConfig(
                pool_configs=[ipam_pool]
            )
            self.docker_network = self.docker.networks.create(
                self.name,
                driver='bridge',
                ipam=ipam_config
            )

    def down(self, destroy=False):
        """
        Ensures that the network is down.
        """
        if destroy and self.is_up():
            with log.step("Removing network"):
                self.docker_network.remove()
                self.docker_network = None

    def _refresh_network_status(self):
        try:
            self.docker_network = self.docker.networks.get(self.name)
        except docker_lib.errors.NotFound:
            self.docker_network = None


class IPPool(object):
    """
    Manages IP addresses for all containers in the network.
    """
    def __init__(self):
        self.subnet = ip_network(NETWORK_SUBNET)
        self.iprange = ip_network(NETWORK_IPRANGE)
        self.gateway = ip_address(NETWORK_GATEWAY)
        self.name_to_ip = {}
        self.ip_to_name = {}

    def add(self, container):
        """
        Adds the container to the pool, returns a string representation of the
        assigned IP address.
        """
        address = self._get_next_free_ip()
        self.name_to_ip[container.name] = address
        self.ip_to_name[address] = container.name
        return str(address)

    def get_ip_for_container_name(self, name):
        """
        Get the string representation of the IP address for a given container
        name, None if does not exist.
        """
        return str(self.name_to_ip.get(name))

    def get_container_name_for_ip(self, ip):
        """
        Get the name of the container for a given IP address.
        """
        return self.ip_to_name.get(ip_address(ip))

    def _get_next_free_ip(self):
        for address in self.subnet.hosts():
            if address not in self.ip_to_name.keys() \
                and address not in self.iprange \
                    and address != self.gateway:
                return address

        log.fatal("the IP address pool is too small to assign another IP,"
                  "change the network settings or the number of nodes")
        return None
