"""
Module for reading and validating the network configuration file.
"""
from collections import namedtuple
from schema import Schema, SchemaError, Optional, Or
import yaml

from nettool import log


NODE = Schema({
    "name": str,
    Optional("seed"): bool,
    Optional("role"): Or("follower", "leader", "audit"),
    Optional("ui_port"): int
})


RULE = Schema({
    "action": Or("allow", "deny"),
    Optional("source"): str,
    Optional("target"): str,
    Optional("one-way"): bool
})


CONFIG = Schema({
    "nodes": [NODE],
    "network": {
        "rules": [RULE]
    },
})


Environment = namedtuple("Environment", "nodes, network")

Node = namedtuple("Node", "name, seed, role, ui_port")

Network = namedtuple("Network", "rules")

Rule = namedtuple("Rule", "source, target, action")


def read_file(config_path):
    """
    Reads the network setup from the config file.
    """
    cfg = _read_yaml(config_path)
    _validate_schema(cfg)
    return _parse_env_config(cfg)


def _read_yaml(path):
    with open(path) as net_file:
        return yaml.load(net_file)


def _validate_schema(cfg):
    try:
        CONFIG.validate(cfg)
    except SchemaError as exc:
        log.fatal(exc)


def _parse_env_config(cfg):
    return Environment(
        nodes=[_parse_node(node) for node in cfg["nodes"]],
        network=_parse_network(cfg["network"])
    )


def _parse_node(cfg):
    return Node(
        name=cfg["name"],
        seed=cfg.get("seed", False),
        role=cfg.get("role", "follower"),
        ui_port=cfg.get("ui_port", None))


def _parse_network(cfg):
    rules = []

    for rule_cfg in cfg["rules"]:
        source = rule_cfg.get("source", "*")
        target = rule_cfg.get("target", "*")
        action = rule_cfg.get("action", "deny")
        one_way = rule_cfg.get("one-way", False)

        rules.append(Rule(source, target, action))
        if not one_way:
            rules.append(Rule(target, source, action))

    return Network(rules=rules)
