"""
Helper functions for dealing with docker libraries.
"""
import time
from enum import Enum
from termcolor import colored

import requests
import docker as docker_lib

from nettool import log


TRANSITION_WAIT_TIME = 1   # seconds
MAX_TRANSITION_WAITS = 10  # will wait for max 10 * TRANSITION_WAIT_TIME


def create():
    """
    Initialize docker client and verify that we can successfully connect to it.
    """
    with log.step("Connecting to docker"):
        client = docker_lib.from_env()
        _verify_docker_connectivity(client)
        return client


def _verify_docker_connectivity(client):
    try:
        client.ping()
    except requests.exceptions.ConnectionError:
        log.fatal(
            "cannot connect to the docker daemon,",
            "make sure it is running and that you can connect to it."
        )


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


class Image(object):
    """
    A wrapper for the docker image.
    """
    def __init__(self, docker, *, tag, path):
        self.docker = docker
        self.tag = tag
        self.path = path
        self.docker_image = None

    @property
    def is_built(self):
        """
        Retrieves the current status of the docker image corresponding to this
        service and returns True if it was already built.

        Note that this does not mean that the image is up to date, only that it
        exists.
        """
        self._refresh_status()
        return self.docker_image is not None

    def build(self, rebuild=False):
        """
        Builds the docker image for the service unless it already exists. If
        the rebuild parameter is set to True, builds the image even if it
        exists.
        """
        if not self.is_built or rebuild:
            with log.step("Building image", self.tag):
                self.docker_image = self.docker.images.build(
                    path=self.path,
                    tag=self.tag,
                    rm=True
                )

    def destroy(self):
        """
        Destroys the docker image for the service, unless it was already
        destroyed.
        """
        if self.is_built:
            with log.step("Removing image", self.tag):
                self.docker.images.remove(self.tag, force=True)
                self.docker_image = None
        else:
            log.info("Image", self.tag, "already removed")

    def _refresh_status(self):
        try:
            self.docker_image = self.docker.images.get(self.tag)
        except docker_lib.errors.ImageNotFound:
            self.docker_image = None


class Container(object):
    """
    A docker container for a service. Base class providing functionality for
    all service container classes.
    """
    def __init__(self, docker, *, name, image, extra_args=None):
        self.docker = docker
        self.name = name
        self.image = image
        self.docker_container = None
        self.extra_args = extra_args or {}
        self.assigned_ip = None
        self.network = None

    @property
    def status(self):
        """
        Retrieves the current status of the docker container corresponding to
        this service instance and returns True if it is running.
        """
        self._refresh_status()
        if not self.docker_container:
            return Status.DELETED
        return Status(self.docker_container.status)

    @property
    def is_running(self):
        """
        Checks if container is currently running.
        """
        return self.status == Status.RUNNING

    def print_info(self):
        """
        Prints the status of the docker container corresponding to this
        instance.
        """
        if self.is_running:
            log.info("Container status:", colored("UP", "green"))
        else:
            log.info("Container status:", colored("DOWN", "red"))
        log.info("Container name:", self.name)
        log.info("Image tag:", self.image.tag)
        if self.assigned_ip:
            log.info("Assigned IP:", self.assigned_ip)
        self._print_actual_network_info()

    def up(self, restart=False):
        """
        Ensures that the docker container corresponding to this service is
        running. If the restart parameter is set to true, the container is
        restarted from scratch.
        """
        self.image.build()

        if restart:
            self.down(destroy=True)

        with log.step("Starting", self.name):
            self._wait_for_transition()

            if self.status == Status.DELETED:
                self._create_container()

            if self.status in [Status.CREATED, Status.EXITED]:
                self._connect_to_network()
                self.docker_container.start()

            if self.status == Status.PAUSED:
                self.docker_container.unpause()

            if self.status == Status.RUNNING:
                return

            if self.status == Status.DEAD:
                log.fatal("Container", self.name, "is in a dead state")

    def down(self, destroy=False):
        """
        Ensures that the docker container corresponding to this service is
        currently not running:
         - if the container is running, stops it
         - if the destroy parameter is set to true, also removes the container.
        """
        with log.step("Stopping", self.name):
            self._wait_for_transition()

            if self.status in [Status.RUNNING, Status.PAUSED]:
                self.docker_container.stop()
                self._disconnect_from_networks()
                if not destroy:
                    return

            if self.status in [Status.CREATED, Status.EXITED]:
                self.docker_container.remove(v=True)

            self._wait_for_transition()

            if self.status == Status.DELETED:
                return

            if self.status == Status.DEAD:
                log.fatal("Container", self.name, "is in a dead state")

    def _create_container(self):
        kwargs = self.extra_args.copy()
        kwargs["name"] = self.name
        kwargs["hostname"] = self.name
        kwargs["detach"] = True

        if self.network:
            kwargs["network"] = self.network.name

        self.docker_container = self.docker.containers.create(
            self.image.tag,
            **kwargs
        )

    def _refresh_status(self):
        try:
            self.docker_container = self.docker.containers.get(self.name)
        except docker_lib.errors.NotFound:
            self.docker_container = None

    def _wait_for_transition(self):
        for _ in range(MAX_TRANSITION_WAITS):
            if Status.is_transition(self.status):
                time.sleep(TRANSITION_WAIT_TIME)
            else:
                break
        else:
            log.fatal("Timeout when waiting for the container",
                      self.name, "to move to another state")

    def _connected_networks(self):
        net_info = self.docker_container.attrs["NetworkSettings"]["Networks"]
        network_names = [key for key, _ in net_info.items()]
        return [self.docker.networks.get(name) for name in network_names]

    def _disconnect_from_networks(self):
        if not self.network:
            return

        for network in self._connected_networks():
            network.disconnect(self.docker_container)

    def _connect_to_network(self):
        if not self.network:
            return

        self._disconnect_from_networks()

        self.network.docker_network.connect(
            self.docker_container,
            ipv4_address=self.assigned_ip
        )

    def _print_actual_network_info(self):
        if self.is_running:
            attrs = self.docker_container.attrs
            net_info = attrs["NetworkSettings"]["Networks"]
            log.info("Networks: ")
            for net_name, net_attrs in net_info.items():
                net_ip = net_attrs.get("IPAddress")
                log.info(" - ", net_name + ":", net_ip)
