"""
Helper functions for dealing with docker libraries.
"""
import requests
import docker

from nettool import log


def create():
    """
    Initialize docker client and verify that we can successfully connect to it.
    """
    with log.step("Connecting to docker"):
        client = docker.from_env()
        _verify_docker_connectivity(client)
        return client


def _verify_docker_connectivity(client):
    try:
        client.ping()
    except requests.exceptions.ConnectionError:
        log.fatal(
            "cannot connect to a docker instance,",
            "make sure it is running and that you can connect to it."
        )
