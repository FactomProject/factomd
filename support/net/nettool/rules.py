"""
Module for manipulating the iptables rules. All rules are created by issuing
iptables commands in the gateway container and appended to a custom chain that
is injected into the DOCKER-USER chain.
"""
from enum import Enum
from ipaddress import ip_network
import re

from nettool import log, services


# Custom chain that we inject rules to
IPTABLES_CHAIN = "FACTOMD-NETTOOL"


# Regex for parsing iptables output
RULE_REGEX = re.compile(
    "".join([
        "-A %s" % IPTABLES_CHAIN,
        r" -s (?P<source>\S+)",
        r" -d (?P<target>\S+)",
        r" -j (?P<action>\S+)"
    ]),
    re.IGNORECASE
)


# Iptables commands
LIST_CUSTOM_RULES = \
    "iptables -S " + IPTABLES_CHAIN
CREATE_CUSTOM_CHAIN = \
    "iptables -N " + IPTABLES_CHAIN
DELETE_CUSTOM_CHAIN = \
    "iptables -X " + IPTABLES_CHAIN
CREATE_FORWARD_TO_CUSTOM_CHAIN = \
    "iptables -I DOCKER-USER -j " + IPTABLES_CHAIN
CHECK_FORWARD_TO_CUSTOM_CHAIN = \
    "iptables -C DOCKER-USER -j " + IPTABLES_CHAIN
DELETE_FORWARD_TO_CUSTOM_CHAIN = \
    "iptables -D DOCKER-USER -j " + IPTABLES_CHAIN
FLUSH_CUSTOM_CHAIN = \
    "iptables -F " + IPTABLES_CHAIN


class Rules(object):
    """
    Allows managing iptables rules via the gateway container.
    """
    env = None

    def __init__(self, docker, config, network, testnet):
        self.network = network
        self.testnet = testnet
        self.gateway = services.Gateway(docker)
        self.rules = [Rule.from_cfg(rule) for rule in config.rules]

    def print_info(self):
        """
        Prints the information about the status of iptables rules.
        """
        self.gateway.print_info()
        parsed, unparsed = self._parse_rules()

        if parsed:
            log.info("Rules:")
            for rule in parsed:
                log.info("   ", rule)

        if unparsed:
            log.info("Unmanaged rules:")
            for rule in unparsed:
                log.info("  ", rule)

    def up(self, build=False):
        """
        Creates the custom chain and adds all the initial rules.
        """
        if build:
            services.Gateway.rebuild_image()
        self.gateway.up(restart=build)
        with log.step("Creating iptables rules"):
            self.gateway.run(CREATE_CUSTOM_CHAIN)
            result = self.gateway.run(CHECK_FORWARD_TO_CUSTOM_CHAIN)
            if result:
                self.gateway.run(CREATE_FORWARD_TO_CUSTOM_CHAIN)
            self.gateway.run(FLUSH_CUSTOM_CHAIN)
            for rule in self.rules:
                if not self._rule_exists(rule):
                    self.gateway.run(rule.append_cmd(self.network))

    def down(self, destroy=False):
        """
        Destroys the custom chain and all its rules.
        """
        with log.step("Deleting iptables rules"):
            self.gateway.run(FLUSH_CUSTOM_CHAIN)
            self.gateway.run(DELETE_FORWARD_TO_CUSTOM_CHAIN)
            self.gateway.run(DELETE_CUSTOM_CHAIN)
            self.gateway.down(destroy=destroy)

    def insert(self, source, target, action, one_way):
        """
        Insert a rule at the beginning of the custom chain.
        """
        self._ensure_gateway()
        action = RuleAction.from_cfg(action)
        self._insert_rule(Rule(source, target, action))
        if not one_way:
            self._insert_rule(Rule(target, source, action))

    def append(self, source, target, action, one_way):
        """
        Adds a rule to the custom chain.
        """
        self._ensure_gateway()
        action = RuleAction.from_cfg(action)
        self._append_rule(Rule(source, target, action))
        if not one_way:
            self._append_rule(Rule(target, source, action))

    def delete(self, source, target, action, one_way):
        """
        Deletes the rule from the custom chain.
        """
        self._ensure_gateway()
        action = RuleAction.from_cfg(action)
        self._delete_rule(Rule(source, target, action))
        if not one_way:
            self._delete_rule(Rule(target, source, action))

    def _insert_rule(self, rule):
        if self._rule_exists(rule):
            log.info("Rule", repr(rule), "already exists")
        else:
            self.gateway.run(rule.insert_cmd(self.network))
            log.info("Inserted new rule:", repr(rule))

    def _append_rule(self, rule):
        if self._rule_exists(rule):
            log.info("Rule", repr(rule), "already exists")
        else:
            self.gateway.run(rule.append_cmd(self.network))
            log.info("Appended new rule:", repr(rule))

    def _delete_rule(self, rule):
        if self._rule_exists(rule):
            self.gateway.run(rule.delete_cmd(self.network))
            log.info("Deleted rule:", repr(rule))
        else:
            log.info("Rule", repr(rule), "not found")

    def _rule_exists(self, rule):
        # iptables result empty -> rule exists
        return not self.gateway.run(rule.check_cmd(self.network))

    def _parse_rules(self):
        if not self.gateway.is_running:
            return None, None

        lines = self.gateway.run(LIST_CUSTOM_RULES).splitlines()

        parsed = []
        unparsed = []

        for line in lines[1:]:  # skip chain creation rule
            rule = Rule.parse(line, self.network)
            if isinstance(rule, Rule):
                parsed.append(rule)
            else:
                unparsed.append(rule)

        return parsed, unparsed

    def _ensure_gateway(self):
        if not self.gateway.is_running:
            log.fatal("Gateway container must be up to change rules")


class RuleAction(Enum):
    """
    An action for the rule.
    """
    ACCEPT = "ACCEPT"
    DROP = "DROP"

    @staticmethod
    def from_cfg(cfg):
        """
        Create a rule action from the parsed config file.
        """
        if cfg == "allow":
            return RuleAction.ACCEPT
        elif cfg == "deny":
            return RuleAction.DROP
        else:
            raise Exception("Unrecognized action: " + cfg)

    def __repr__(self):
        if self == RuleAction.ACCEPT:
            return "allow"
        elif self == RuleAction.DROP:
            return "deny"
        else:
            raise Exception("Unrecognized action: " + self)


class Rule(object):
    """
    Represents a single managed iptables rule.
    """
    @staticmethod
    def from_cfg(cfg):
        """
        Creates a rule from a parsed config file.
        """
        action = RuleAction.from_cfg(cfg.action)
        return Rule(cfg.source, cfg.target, action)

    @staticmethod
    def parse(line, network):
        """
        Attempts to parse an iptables output line into a rule, returns the rule
        if parsing was successful, otherwise returns the original line.
        """
        match = RULE_REGEX.match(line)
        if not match:
            return line

        ip_source = match.group('source')
        ip_target = match.group('target')
        ip_action = match.group('action')

        source = _parse_network_address(ip_source, network)
        target = _parse_network_address(ip_target, network)

        return Rule(source, target, RuleAction(ip_action))

    def __init__(self, source, target, action):
        self.source = source
        self.target = target
        self.action = action

    def insert_cmd(self, network):
        """
        Returns an insert command for the current rule.
        """
        ip_source = _name_to_ip(self.source, network)
        ip_target = _name_to_ip(self.target, network)
        return "iptables -I " + IPTABLES_CHAIN + \
            " -s " + ip_source + \
            " -d " + ip_target + \
            " -j " + self.action.value

    def append_cmd(self, network):
        """
        Returns a create command for the current rule.
        """
        ip_source = _name_to_ip(self.source, network)
        ip_target = _name_to_ip(self.target, network)
        return "iptables -A " + IPTABLES_CHAIN + \
            " -s " + ip_source + \
            " -d " + ip_target + \
            " -j " + self.action.value

    def check_cmd(self, network):
        """
        Returns a check command for the current rule.
        """
        ip_source = _name_to_ip(self.source, network)
        ip_target = _name_to_ip(self.target, network)
        return "iptables -C " + IPTABLES_CHAIN + \
            " -s " + ip_source + \
            " -d " + ip_target + \
            " -j " + self.action.value

    def delete_cmd(self, network):
        """
        Returns a delete command for the current rule.
        """
        ip_source = _name_to_ip(self.source, network)
        ip_target = _name_to_ip(self.target, network)
        return "iptables -D " + IPTABLES_CHAIN + \
            " -s " + ip_source + \
            " -d " + ip_target + \
            " -j " + self.action.value

    def __repr__(self):
        return "".join([
            repr(self.action),
            ": ",
            self.source,
            " -> ",
            self.target
        ])


def _parse_network_address(string, network):
    net = ip_network(string)
    if net.num_addresses == 1:  # "/32"
        ip = str(net.network_address)
        name = network.ip_pool.get_container_name_for_ip(ip)
        if name:
            return name
        return ip

    if net == network.address:
        return "*"

    return str(net)


def _name_to_ip(name, network):
    if name == "*":
        return str(network.address)
    container_ip = network.ip_pool.get_ip_for_container_name(name)
    if container_ip:
        return str(container_ip)

    return name
