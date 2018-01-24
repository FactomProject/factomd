"""
Module for manipulating the docker network that allows connectivity between
nodes.
"""
from ipaddress import ip_address, ip_network
import docker as docker_lib

from nettool import log


NETWORK_NAME = "nettool"
NETWORK_SUBNET = "10.12.0.0/24"
NETWORK_GATEWAY = "10.12.0.254"
NETWORK_IPRANGE = "10.12.1.0/24"


class Network(object):
    """
    A docker network that takes care of static IP assignment and allows
    connectivity between containers.
    """
    env = None

    def __init__(self, containers):
        self.name = NETWORK_NAME
        self.docker_network = None
        self.ip_pool = IPPool()

        self.containers = []
        self._assign_ips(containers)

    @property
    def address(self):
        """
        Get the address of the network.
        """
        return ip_network(NETWORK_SUBNET)

    def print_info(self):
        """
        Print information about the network.
        """
        log.section("Network")
        log.info("Name:", self.name)

    def is_up(self):
        """
        Checks if the docker network is up.
        """
        self._refresh_network_status()
        return self.docker_network is not None

    def up(self):
        """
        Ensures that the network is running.
        """
        if self.is_up():
            return

        with log.step("Creating network"):
            ipam_pool = docker_lib.types.IPAMPool(
                subnet=self.ip_pool.get_subnet(),
                gateway=self.ip_pool.get_network_gateway(),
                iprange=self.ip_pool.get_iprange()
            )
            ipam_config = docker_lib.types.IPAMConfig(
                pool_configs=[ipam_pool]
            )
            self.docker_network = self.env.docker.networks.create(
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
            self.docker_network = self.env.docker.networks.get(self.name)
        except docker_lib.errors.NotFound:
            self.docker_network = None

    def _assign_ips(self, containers):
        for container in containers:
            if container.in_network:
                self.containers.append(container)
                ip = self.ip_pool.add(container)
                container.ip_address = ip


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
        name = container.instance_name
        address = self._get_next_free_ip()
        self.name_to_ip[name] = address
        self.ip_to_name[address] = name
        return str(address)

    def get_subnet(self):
        """
        Get the string representation of the subnet.
        """
        return str(self.subnet)

    def get_network_gateway(self):
        """
        Get the string representation of the IP address of the network gateway.
        """
        return str(self.gateway)

    def get_iprange(self):
        """
        Get the string representation of the IP range.
        """
        return str(self.iprange)

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
