"""
Provides classes that describe services used in the network setup. Each service
class corresponds to a single docker image and each instance of the class to
a docker container.
"""
import os.path

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
    CTX_PATH = "docker/gateway"

    def __init__(self):
        super().__init__()
        self.in_network = False

    @classmethod
    def _build_image(cls):
        return cls.env.docker.images.build(
            path=cls.CTX_PATH,
            tag=cls.IMAGE_TAG,
            rm=True
        )

    def _create_container(self):
        return self.env.docker.containers.create(
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

    def print_info(self):
        log.section("Gateway")
        self.print_container_info()

    def exec_run(self, cmd):
        """
        Execute a custom command on the gateway.
        """
        if not self.is_running:
            return ""

        return self.container.exec_run(cmd).decode("ascii")


class SeedServer(Container):
    """
    The seeds server is a container that runs an nginx instance and serves the
    list of factomd nodes for discovery.
    """
    NAME = "seeds_server"
    IMAGE_TAG = "nettool_nginx"
    CTX_PATH = "docker/seeds"
    SEEDS_FILE_LOCAL = "docker/seeds/seeds"
    SEEDS_FILE_REMOTE = "/usr/share/nginx/html/seeds"

    @classmethod
    def _build_image(cls):
        return cls.env.docker.images.build(
            path=cls.CTX_PATH,
            tag=cls.IMAGE_TAG,
            rm=True
        )

    def __init__(self, nodes):
        super().__init__()
        self.seed_nodes = [node for node in nodes if node.is_seed]

    def _create_container(self):
        return self.env.docker.containers.create(
            self.IMAGE_TAG,
            name=self.instance_name,
            hostname=self.instance_name,
            network=self.env.network.name,

            volumes={
                os.path.abspath(self.SEEDS_FILE_LOCAL): {
                    'bind': self.SEEDS_FILE_REMOTE,
                    'mode': 'ro'
                }
            },
            detach=True
        )

    def print_info(self):
        log.section("Seeds server")

        self.print_container_info()
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

    def generate_seeds_file(self):
        """
        Write all currently know seeds to a file that is mapped to a file in
        the container, so that the server seed list gets updated.
        """
        entries = [n.seed_entry for n in self.seed_nodes]
        with open(self.SEEDS_FILE_LOCAL, "w") as seed_file:
            seed_file.writelines(entries)


class Factomd(Container):
    """
    A factomd instance.
    """
    NAME = "factomd"
    IMAGE_TAG = "nettool_factomd"
    CTX_PATH = "../../"

    @classmethod
    def _build_image(cls):
        return cls.env.docker.images.build(
            path=cls.CTX_PATH,
            tag=cls.IMAGE_TAG,
            rm=True
        )

    def __init__(self, config):
        super().__init__()
        self.config = config

    @property
    def instance_name(self):
        return self.config.name

    @property
    def is_seed(self):
        """
        If True, this node will be added to the list of seeds.
        """
        return self.config.seed

    @property
    def server_port(self):
        """
        Gets the port for communication between servers.
        """
        return self.config.server_port

    @property
    def seed_entry(self):
        """
        Gets a string for an entry in the seed list for this node.
        """
        return f"{self.ip_address}:8110"

    def _create_container(self):
        return self.env.docker.containers.create(
            self.IMAGE_TAG,
            name=self.instance_name,
            hostname=self.instance_name,
            network=self.env.network.name,
            detach=True
        )

    def print_info(self):
        log.section("Node", self.config.name)
        self.print_container_info()
