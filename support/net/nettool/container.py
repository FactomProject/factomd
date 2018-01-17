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

    image = None
    container = None

    @classmethod
    def is_built(cls, docker):
        """
        Retrieves the current status of the docker image corresponding to this
        service and returns True if it was already built.

        Note that this does not mean that the image is up to date, only that it
        exists.
        """
        cls._refresh_image_status(docker)
        return cls.image is not None

    @classmethod
    def build(cls, docker, rebuild=False):
        """
        Builds the docker image for the service unless it already exists. If
        the rebuild parameter is set to True, builds the image even if it
        exists.
        """
        if not cls.is_built(docker) or rebuild:
            with log.step("Building image", cls.IMAGE_TAG):
                cls.image = cls._build_image(docker)

    @classmethod
    def destroy(cls, docker):
        """
        Destroys the docker image for the service, unless it was already
        destroyed.
        """
        if cls.is_built(docker):
            with log.step("Removing image", cls.IMAGE_TAG):
                cls.image = docker.images.remove(cls.IMAGE_TAG, force=True)
        else:
            log.info("Image", cls.IMAGE_TAG, "already removed")


    @classmethod
    @abstractmethod
    def _build_image(cls, docker):
        """
        Builds the docker image for the current service. To be overwritten in
        inherited classes.
        """
        pass

    @property
    def instance_name(self):
        """
        Name of the container.
        """
        return self.NAME

    def print_container_info(self, docker):
        """
        Prints the status of the docker container corresponding to this
        instance.
        """
        if self.status(docker) == Status.RUNNING:
            log.info("Container status:", colored("UP", "green"))
        else:
            log.info("Container status:", colored("DOWN", "red"))
        log.info("Container name:", self.instance_name)
        log.info("Image tag:", self.IMAGE_TAG)


    def status(self, docker):
        """
        Retrieves the current status of the docker container corresponding to
        this service instance and returns True if it is running.
        """
        self._refresh_container_status(docker)
        if not self.container:
            return Status.DELETED

        return Status(self.container.status)

    def up(self, docker, restart=False):
        """
        Ensures that the docker container corresponding to this service is
        currently running:
         - builds the docker image if it doesn't exist yet
         - checks if the container is in the running state
         - if the container is not running, starts it.

        If the restart parameter is set to true, the container is always
        restarted from scratch.
        """
        self.__class__.build(docker)

        if restart:
            self.down(docker, destroy=True)

        with log.step("Starting", self.instance_name):
            self._wait_for_transition(docker)

            if self.status(docker) == Status.DELETED:
                self._run_container(docker)

            if self.status(docker) in [Status.CREATED, Status.EXITED]:
                self.container.start()

            if self.status(docker) == Status.PAUSED:
                self.container.unpause()

            if self.status(docker) == Status.RUNNING:
                return

            if self.status(docker) == Status.DEAD:
                log.fatal("Container", self.instance_name,
                          "is in a dead state")


    def down(self, docker, destroy=False):
        """
        Ensures that the docker container corresponding to this service is
        currently not running:
         - if the container is running, stops it
         - if the destroy parameter is set to true, also removes the container.
        """
        with log.step("Stopping", self.instance_name):
            self._wait_for_transition(docker)

            if self.status(docker) in [Status.RUNNING, Status.PAUSED]:
                self.container.stop()
                if not destroy:
                    return

            if self.status(docker) in [Status.CREATED, Status.EXITED]:
                self.container.remove(v=True)

            self._wait_for_transition(docker)

            if self.status(docker) == Status.DELETED:
                return

            if self.status(docker) == Status.DEAD:
                log.fatal("Container", self.instance_name,
                          "is in a dead state")

    @abstractmethod
    def print_info(self, docker):
        """
        Prints the status of this service instance. To be overwritten in the
        inherited classes.
        """
        pass

    @abstractmethod
    def _run_container(self, docker):
        pass

    @classmethod
    def _refresh_image_status(cls, docker):
        try:
            cls.image = docker.images.get(cls.IMAGE_TAG)
        except docker_lib.errors.ImageNotFound:
            cls.image = None

    def _refresh_container_status(self, docker):
        try:
            self.container = docker.containers.get(self.instance_name)
        except docker_lib.errors.NotFound:
            self.container = None

    def _wait_for_transition(self, docker):
        for _ in range(MAX_TRANSITION_WAITS):
            if Status.is_transition(self.status(docker)):
                time.sleep(TRANSITION_WAIT_TIME)
            else:
                break
        else:
            log.fatal("Timeout when waiting for the container",
                      self.instance_name, "to move to another state")
