"""
Definitions of various services run in the environment.
"""
import os.path
import time

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
        return self.container.docker_container.exec_run(cmd)


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
            seed_file.write("\n".join(entries))


class Factomd(Service):
    """
    A factomd instance.
    """
    WAIT_FOR_V2_API_TIMEOUT_SECS = 120

    def __init__(self, docker, config, identity, flags):
        super().__init__()

        self.identity = identity

        self.__class__.image = Image(
            docker,
            tag="nettool_factomd",
            path="docker/node",
        )

        extra_args = {
            "environment": {
                "ID_CHAIN": self.identity.chain_id,
                "PRIV_KEY": self.identity.priv_key,
                "PUB_KEY": self.identity.pub_key
            },
            "ports": {}
        }

        if config.ui_port:
            extra_args["ports"]["8090"] = config.ui_port

        if config.api_port:
            extra_args["ports"]["8088"] = config.api_port

        if flags:
            extra_args["environment"]["FLAGS"] = flags

        self.container = Container(
            docker,
            image=self.image,
            name=config.name,
            extra_args=extra_args
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
    def seed_entry(self):
        """
        Gets a string for an entry in the seed list for this node.
        """
        return f"{self.container.assigned_ip}:8110"

    def up(self, restart=False):
        """
        Bring the service up, if "restart" is True, the service container is
        restarted.
        """
        super().up(restart=restart)

        if self.is_leader or self.is_audit:
            self._wait_for_api()
            self._promote()

    def print_info(self):
        """
        Prints the current status of the node.
        """
        log.section("Node", self.config.name)
        log.info("Role:", self.config.role)
        self.container.print_info()

    def _wait_for_api(self):
        cmd = f"wait_for_port.sh 8088 {self.WAIT_FOR_V2_API_TIMEOUT_SECS}"
        with log.step(f"Waiting for {self.instance_name} API"):
            result, output = self._run(cmd)
            if result != 0:
                log.fatal("Failed:", output)

    def _promote(self):
        if self.is_leader:
            server_type = "f"
        elif self.is_audit:
            server_type = "a"
        else:
            log.fatal("Unknown server role")

        cmd = " ".join([
            "addservermessage",
            "-host=localhost:8088",
            "send",
            server_type,
            self.identity.chain_id,
            self.identity.priv_key
        ])

        result, output = self._run(cmd)
        if result != 0:
            log.fatal("Failed to promote", self.instance_name, output)

    def _run(self, cmd):
        if not self.container.is_running:
            log.fatal("Container startup failed")
        return self.container.docker_container.exec_run(cmd)
