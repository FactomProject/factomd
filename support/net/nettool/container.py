"""
Base library for managing service containers.
"""
from enum import Enum
from abc import ABC, abstractmethod
import time

from termcolor import colored
import docker as docker_lib

from nettool import log


TRANSITION_WAIT_TIME = 1   # seconds
MAX_TRANSITION_WAITS = 10  # will wait for max 10 * TRANSITION_WAIT_TIME


class Status(Enum):
    """
    Possible container statuses.
    """
    CREATED = "created"
    RESTARTING = "restarting"
    RUNNING = "running"
    REMOVING = "removing"
    PAUSED = "paused"
    EXITED = "exited"
    DEAD = "dead"
    DELETED = "deleted"

    @staticmethod
    def is_transition(status):
        """
        Indicates if we should wait for the container to exit temporary state.
        """
        return status in [Status.RESTARTING, Status.REMOVING]


class Container(ABC):
    """
    A docker container for a service. Base class providing functionality for
    all service container classes.
    """
    NAME = "unknown"
    IMAGE_TAG = "unknown"

    env = None
    image = None
    container = None

    @classmethod
    def is_built(cls):
        """
        Retrieves the current status of the docker image corresponding to this
        service and returns True if it was already built.

        Note that this does not mean that the image is up to date, only that it
        exists.
        """
        cls._refresh_image_status()
        return cls.image is not None

    @classmethod
    def build(cls, rebuild=False):
        """
        Builds the docker image for the service unless it already exists. If
        the rebuild parameter is set to True, builds the image even if it
        exists.
        """
        if not cls.is_built() or rebuild:
            with log.step("Building image", cls.IMAGE_TAG):
                cls.image = cls._build_image()

    @classmethod
    def destroy(cls):
        """
        Destroys the docker image for the service, unless it was already
        destroyed.
        """
        if cls.is_built():
            with log.step("Removing image", cls.IMAGE_TAG):
                docker = cls.env.docker
                cls.image = docker.images.remove(cls.IMAGE_TAG, force=True)
        else:
            log.info("Image", cls.IMAGE_TAG, "already removed")

    @classmethod
    @abstractmethod
    def _build_image(cls):
        """
        Builds the docker image for the current service. To be overwritten in
        inherited classes.
        """
        pass

    def __init__(self):
        self.ip_address = None
        self.in_network = True

    @property
    def instance_name(self):
        """
        Name of the container.
        """
        return self.NAME

    @property
    def status(self):
        """
        Retrieves the current status of the docker container corresponding to
        this service instance and returns True if it is running.
        """
        self._refresh_container_status()
        if not self.container:
            return Status.DELETED
        return Status(self.container.status)

    @property
    def is_running(self):
        """
        Checks if container is currently running.
        """
        return self.status == Status.RUNNING

    def print_container_info(self):
        """
        Prints the status of the docker container corresponding to this
        instance.
        """
        if self.is_running:
            log.info("Container status:", colored("UP", "green"))
        else:
            log.info("Container status:", colored("DOWN", "red"))
        log.info("Container name:", self.instance_name)
        log.info("Image tag:", self.IMAGE_TAG)
        log.info("Assigned IP:", self.ip_address)
        self._print_actual_network_info()

    def up(self, restart=False):
        """
        Ensures that the docker container corresponding to this service is
        running. If the restart parameter is set to true, the container is
        restarted from scratch.
        """
        self.__class__.build()

        if restart:
            self.down(destroy=True)

        with log.step("Starting", self.instance_name):
            self._wait_for_transition()

            if self.status == Status.DELETED:
                self.container = self._create_container()

            if self.status in [Status.CREATED, Status.EXITED]:
                self._connect_to_network()
                self.container.start()

            if self.status == Status.PAUSED:
                self.container.unpause()

            if self.status == Status.RUNNING:
                return

            if self.status == Status.DEAD:
                log.fatal("Container", self.instance_name,
                          "is in a dead state")

    def down(self, destroy=False):
        """
        Ensures that the docker container corresponding to this service is
        currently not running:
         - if the container is running, stops it
         - if the destroy parameter is set to true, also removes the container.
        """
        with log.step("Stopping", self.instance_name):
            self._wait_for_transition()

            if self.status in [Status.RUNNING, Status.PAUSED]:
                self.container.stop()
                self._disconnect_from_networks()
                if not destroy:
                    return

            if self.status in [Status.CREATED, Status.EXITED]:
                self.container.remove(v=True)

            self._wait_for_transition()

            if self.status == Status.DELETED:
                return

            if self.status == Status.DEAD:
                log.fatal("Container", self.instance_name,
                          "is in a dead state")

    @abstractmethod
    def print_info(self):
        """
        Prints the status of this service instance. To be overwritten in the
        inherited classes.
        """
        pass

    @abstractmethod
    def _create_container(self):
        pass

    @classmethod
    def _refresh_image_status(cls):
        try:
            cls.image = cls.env.docker.images.get(cls.IMAGE_TAG)
        except docker_lib.errors.ImageNotFound:
            cls.image = None

    def _refresh_container_status(self):
        try:
            self.container = self.env.docker.containers.get(self.instance_name)
        except docker_lib.errors.NotFound:
            self.container = None

    def _wait_for_transition(self):
        for _ in range(MAX_TRANSITION_WAITS):
            if Status.is_transition(self.status):
                time.sleep(TRANSITION_WAIT_TIME)
            else:
                break
        else:
            log.fatal("Timeout when waiting for the container",
                      self.instance_name, "to move to another state")

    def _connected_networks(self):
        net_info = self.container.attrs["NetworkSettings"]["Networks"]
        network_names = [key for key, _ in net_info.items()]
        return [self.env.docker.networks.get(name) for name in network_names]

    def _disconnect_from_networks(self):
        if not self.in_network:
            return

        for network in self._connected_networks():
            network.disconnect(self.container)

    def _connect_to_network(self):
        if not self.in_network:
            return

        self._disconnect_from_networks()

        self.env.network.docker_network.connect(
            self.container,
            ipv4_address=self.ip_address
        )

    def _print_actual_network_info(self):
        if self.is_running:
            net_info = self.container.attrs["NetworkSettings"]["Networks"]
            log.info("Networks: ")
            for net_name, net_attrs in net_info.items():
                net_ip = net_attrs.get("IPAddress")
                log.info(" - ", net_name + ":", net_ip)
