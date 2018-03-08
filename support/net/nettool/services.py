"""
Definitions of various services run in the environment.
"""
import os.path

from nettool import log
from nettool.docker_client import Image, Container


class Service(object):
    """
    Base class for all services.
    """
    image = None
    container = None

    @classmethod
    def rebuild_image(cls):
        """
        Rebuild the image for this service.
        """
        cls.image.build(rebuild=True)

    @classmethod
    def destroy_image(cls):
        """
        Destroy the image for this service.
        """
        cls.image.destroy()

    @property
    def is_running(self):
        """
        Check if the service is currently running.
        """
        return self.container.is_running

    def up(self, restart=False):
        """
        Bring the service up, if "restart" is True, the service container is
        restarted.
        """
        self.__class__.image.build()
        self.container.up(restart=restart)

    def down(self, destroy=False):
        """
        Shuts the service down, if "destroy" is True, removes the service
        container.
        """
        self.container.down(destroy=destroy)


class Gateway(Service):
    """
    A gateway is a special container that runs in the privileged mode and
    allows access to the host iptables configuration and the list of processes.

    This allows manipulating iptables settings even if the tool is executed
    inside the Docker for Mac / Windows environment, where the user has no
    direct access to the host VM.
    """
    def __init__(self, docker):
        super().__init__()
        self.__class__.image = Image(
            docker,
            tag="nettool_gateway",
            path="docker/gateway"
        )
        self.container = Container(
            docker,
            image=self.image,
            name="gateway",
            extra_args={
                "network_mode": "host",
                "pid_mode": "host",
                "ipc_mode": "host",
                "privileged": True,
                "stdin_open": True,
                "tty": True,
            }
        )

    def print_info(self):
        """
        Print info for the gateway.
        """
        log.section("Gateway")
        self.container.print_info()

    def run(self, cmd):
        """
        Execute a custom command on the gateway.
        """
        if not self.container.is_running:
            return None
        return self.container.docker_container.exec_run(cmd).decode("ascii")


class SeedServer(Service):
    """
    The seeds server is a container that runs an nginx instance and serves the
    list of factomd nodes for discovery.
    """
    SEEDS_FILE_LOCAL = "docker/seeds/seeds"
    SEEDS_FILE_REMOTE = "/usr/share/nginx/html/seeds"

    def __init__(self, docker, seed_nodes):
        super().__init__()
        self.__class__.image = Image(
            docker,
            tag="nettool_seeds",
            path="docker/seeds"
        )
        self.container = Container(
            docker,
            image=self.image,
            name="seeds_server",
            extra_args={
                "volumes": {
                    os.path.abspath(self.SEEDS_FILE_LOCAL): {
                        'bind': self.SEEDS_FILE_REMOTE,
                        'mode': 'ro'
                    }
                }
            }
        )
        self.seed_nodes = seed_nodes

    def print_info(self):
        """
        Prints the current status of the server.
        """
        log.section("Seeds server")

        self.container.print_info()
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


class Factomd(Service):
    """
    A factomd instance.
    """
    def __init__(self, docker, config):
        super().__init__()
        self.__class__.image = Image(
            docker,
            tag="nettool_factomd",
            path="docker/node"
        )
        self.container = Container(
            docker,
            image=self.image,
            name=config.name
        )
        self.config = config

    @property
    def instance_name(self):
        """
        Returns the name of the instance.
        """
        return self.config.name

    @property
    def is_seed(self):
        """
        If True, this node will be added to the list of seeds.
        """
        return self.config.seed

    @property
    def is_leader(self):
        """
        If True, this node will be promoted to a leader.
        """
        return self.config.role == "leader"

    @property
    def is_audit(self):
        """
        If True, this node will be promoted to an audit server.
        """
        return self.config.role == "audit"

    @property
    def is_follower(self):
        """
        If True, this node will remain a follower.
        """
        return self.config.role == "follower"

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
        return f"{self.container.assigned_ip}:8110"

    def print_info(self):
        """
        Prints the current status of the node.
        """
        log.section("Node", self.config.name)
        log.info("Role:", self.config.role)
        self.container.print_info()
