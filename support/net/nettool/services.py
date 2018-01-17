"""
Provides classes that describe services used in the network setup. Each service
class corresponds to a single docker image and each instance of the class to
a docker container.
"""
from nettool import log
from nettool.container import Container


class Gateway(Container):
    """
    A gateway is a special container that runs in the privileged mode and
    allows access to the host iptables configuration and the list of processes.

    This allows manipulating iptables settings even if the tool is executed
    inside the Docker for Mac / Windows environment, where the user has no
    direct access to the host VM.
    """
    NAME = "gateway"
    IMAGE_TAG = "nettool_gateway"

    @classmethod
    def _build_image(cls, docker):
        return docker.images.build(
            path="docker/gateway",
            tag=cls.IMAGE_TAG,
            rm=True
        )

    def _start_container(self, docker):
        return docker.containers.run(
            self.IMAGE_TAG,
            name=self.instance_name,
            hostname=self.instance_name,
            network_mode="host",
            pid_mode="host",
            ipc_mode="host",
            privileged=True,
            stdin_open=True,
            tty=True,
            detach=True
        )

    def print_status(self, docker):
        log.section("Gateway")
        self.print_container_status(docker)


class SeedServer(Container):
    """
    The seeds server is a container that runs an nginx instance and serves the
    list of factomd nodes for discovery.
    """
    NAME = "seeds_server"
    IMAGE_TAG = "nettool_nginx"

    @classmethod
    def _build_image(cls, docker):
        return docker.images.build(
            path="docker/seeds",
            tag=cls.IMAGE_TAG,
            rm=True
        )

    def __init__(self):
        self.seed_nodes = []

    def _start_container(self, docker):
        pass

    def print_status(self, docker):
        log.section("Seeds server")
        self.print_container_status(docker)
        log.info()

        if not self.seed_nodes:
            log.info("Seed nodes: none")
        else:
            log.info("Seed nodes:")
            for node in self.seed_nodes:
                log.info(" -", node.instance_name)

    def add(self, node):
        """
        Add the seed node, so that it can be discovered by other nodes.
        """
        self.seed_nodes.append(node)


class Factomd(Container):
    """
    A factomd instance which is part of the network.
    """
    NAME = "factomd"
    IMAGE_TAG = "nettool_factomd"

    @classmethod
    def _build_image(cls, docker):
        return docker.images.build(
            path="../../",
            tag=cls.IMAGE_TAG,
            rm=True
        )

    def __init__(self, config):
        self.config = config

    @property
    def instance_name(self):
        return self.config.name

    def _start_container(self, docker):
        return docker.containers.run(
            self.IMAGE_TAG,
            name=self.instance_name,
            hostname=self.instance_name,
            detach=True
        )

    def print_status(self, docker):
        log.section("Node", self.config.name)
        self.print_container_status(docker)
