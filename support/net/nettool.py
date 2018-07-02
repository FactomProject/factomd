#!/usr/bin/env python3

"""
A tool for testing a factomd network in a docker environment.

Usage:
    ./nettool.py (-h | --help)
    ./nettool.py status [-f FILE | --file FILE]
    ./nettool.py up [--build] [-f FILE | --file FILE]
    ./nettool.py down [--destroy] [-f FILE | --file FILE]
    ./nettool.py ins <from> <to> <action> [--one-way] [-f FILE | --file FILE]
    ./nettool.py add <from> <to> <action> [--one-way] [-f FILE | --file FILE]
    ./nettool.py del <from> <to> <action> [--one-way] [-f FILE | --file FILE]


Commands:
    status                    Show the current status of the environment.
    up                        Ensure the environment is up and running.
    down                      Stop the environment.
    ins <from> <to> <action>  Insert a new network rule at the beginning.
    add <from> <to> <action>  Append a new network rule at the end.
    del <from> <to> <action>  Delete an existing rule.

Options:
    -h --help               Show this screen.
    -f, --file FILE         YAML config file describing the environment
                            [default: config.yml].
    --build                 Rebuild all container images (useful after e.g.
                            you have updated the factomd code).
    --destroy               Destroy all artifacts created by the tool.
    --one-way               If true, the rule will be added/deleted
                            assymetrically, otherwise the rule is added/deleted
                            both ways.
"""
from docopt import docopt

from nettool import config, environment, log


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
    elif args["ins"]:
        _environment_ins(
            cfg_file,
            args["<from>"],
            args["<to>"],
            args["<action>"],
            args["--one-way"]
        )
    elif args["add"]:
        _environment_add(
            cfg_file,
            args["<from>"],
            args["<to>"],
            args["<action>"],
            args["--one-way"]
        )
    elif args["del"]:
        _environment_del(
            cfg_file,
            args["<from>"],
            args["<to>"],
            args["<action>"],
            args["--one-way"]
        )
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


def _environment_ins(config_file, source, target, action, one_way):
    env = _environment_from_file(config_file)
    env.rules.insert(source, target, action, one_way)


def _environment_add(config_file, source, target, action, one_way):
    env = _environment_from_file(config_file)
    env.rules.append(source, target, action, one_way)


def _environment_del(config_file, source, target, action, one_way):
    env = _environment_from_file(config_file)
    env.rules.delete(source, target, action, one_way)


def _environment_from_file(config_file):
    log.info("Reading config from:", config_file)
    cfg = config.read_file(config_file)
    return environment.Environment(cfg)


if __name__ == "__main__":
    main(docopt(__doc__))
