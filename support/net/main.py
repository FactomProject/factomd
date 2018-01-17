#!/usr/bin/env python

"""
A tool for creating a factomd network in a docker environment.

Usage:
    main.py (-h | --help)
    main.py status [-f FILE | --file FILE]
    main.py up [--build] [-f FILE | --file FILE]
    main.py down [--destroy] [-f FILE | --file FILE]

Commands:
    status                  Show the current status of the network.
    up                      Ensure the network is up and running.
    down                    Stop the network.
    disconnect <from> <to>  Block all connection from factomd node <from> to
                            factomd node <to>. Both <from> and <to> should be
                            names of factomd nodes defined in the network file.
                            By default blocks the connections in both
                            directions, use the --one-way flag to override.
    reconnect <from> <to>   Reconnect previously disconnected <from> and <to>
                            factomd nodes.  By default blocks the connections
                            in both directions, use the --one-way flag to
                            override.

Options:
    -h --help               Show this screen.
    -f, --file FILE         YAML file describing the network
                            [default: network.yml].
    --build                 Rebuild all container images (useful after e.g.
                            you have updated the factomd code).
    --destroy               Destroy all artifacts created by the tool.
    --one-way               Disconnect or reconnect nodes assymetrically,
                            if this option is not specified the connections
                            are dropped / restored both ways.
"""
from docopt import docopt

from nettool import config, docker_client, log, network


def main(args):
    """
    Main entry point for starting the network.
    """
    net_file = args["--file"]
    if args["status"]:
        _print_status(net_file)
    elif args["up"]:
        _network_up(net_file, args["--build"])
    elif args["down"]:
        _network_down(net_file, args["--destroy"])
    else:
        raise Exception("Error parsing arguments")

    log.section("Done")


def _print_status(network_file):
    net = _network_from_file(network_file)
    net.print_status()


def _network_up(network_file, build_mode):
    net = _network_from_file(network_file)
    net.up(build_mode=build_mode)


def _network_down(network_file, destroy_mode):
    net = _network_from_file(network_file)
    net.down(destroy_mode=destroy_mode)


def _network_from_file(network_file):
    docker = docker_client.create()
    log.info("Reading config from:", network_file)
    cfg = config.read_file(network_file)
    return network.Network(cfg, docker)


if __name__ == "__main__":
    main(docopt(__doc__))
