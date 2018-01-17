#!/usr/bin/env python

"""
A tool for creating a factomd network in a docker environment.

Usage:
    ./main.py (-h | --help)
    ./main.py status [-f FILE | --file FILE]
    ./main.py up [--build] [-f FILE | --file FILE]
    ./main.py down [--destroy] [-f FILE | --file FILE]

Commands:
    status                  Show the current status of the environment.
    up                      Ensure the environment is up and running.
    down                    Stop the environment.
    disconnect <from> <to>  Block all connection from factomd node <from> to
                            factomd node <to>. Both <from> and <to> should be
                            names of factomd nodes defined in the config file.
                            By default blocks the connections in both
                            directions, use the --one-way flag to override.
    reconnect <from> <to>   Reconnect previously disconnected <from> and <to>
                            factomd nodes.  By default blocks the connections
                            in both directions, use the --one-way flag to
                            override.

Options:
    -h --help               Show this screen.
    -f, --file FILE         YAML config file describing the environment
                            [default: config.yml].
    --build                 Rebuild all container images (useful after e.g.
                            you have updated the factomd code).
    --destroy               Destroy all artifacts created by the tool.
    --one-way               Disconnect or reconnect nodes assymetrically,
                            if this option is not specified the connections
                            are dropped / restored both ways.
"""
from docopt import docopt

from nettool import config, docker_client, environment, log


def main(args):
    """
    Main entry point for starting the environment.
    """
    cfg_file = args["--file"]
    if args["status"]:
        _print_status(cfg_file)
    elif args["up"]:
        _environment_up(cfg_file, args["--build"])
    elif args["down"]:
        _environment_down(cfg_file, args["--destroy"])
    else:
        raise Exception("Error parsing arguments")

    log.section("Done")


def _print_status(config_file):
    env = _environment_from_file(config_file)
    env.print_info()


def _environment_up(config_file, build_mode):
    env = _environment_from_file(config_file)
    env.up(build_mode=build_mode)


def _environment_down(config_file, destroy_mode):
    env = _environment_from_file(config_file)
    env.down(destroy_mode=destroy_mode)


def _environment_from_file(config_file):
    docker = docker_client.create()
    log.info("Reading config from:", config_file)
    cfg = config.read_file(config_file)
    return environment.Environment(cfg, docker)


if __name__ == "__main__":
    main(docopt(__doc__))
