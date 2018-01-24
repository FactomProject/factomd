"""
Module for manipulating the iptables rules. All rules are created by issuing
iptables commands in the gateway container and appended to a custom chain that
is injected into the DOCKER-USER chain.
"""
from enum import Enum
from ipaddress import ip_network
import re

from nettool import log


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


class Iptables(object):
    """
    Allows managing iptables rules via the gateway container.
    """
    env = None

    def __init__(self, gateway, cfg):
        self.gateway = gateway
        self.rules = [Rule.from_cfg(rule) for rule in cfg.rules]

    def print_info(self):
        """
        Prints the information about the status of iptables rules.
        """
        parsed, unparsed = self._parse_rules()

        if parsed:
            log.info("Rules:")
            for rule in parsed:
                log.info("   ", rule)

        if unparsed:
            log.info("Unmanaged rules:")
            for rule in unparsed:
                log.info("  ", rule)

    def up(self):
        """
        Creates the custom chain and adds all the initial rules.
        """
        with log.step("Creating iptables rules"):
            self._run(CREATE_CUSTOM_CHAIN)
            result = self._run(CHECK_FORWARD_TO_CUSTOM_CHAIN)
            if result:
                self._run(CREATE_FORWARD_TO_CUSTOM_CHAIN)
            self._run(FLUSH_CUSTOM_CHAIN)
            for rule in self.rules:
                rule_exists = not self._run(rule.check_cmd())
                if not rule_exists:
                    self._run(rule.append_cmd())

    def down(self):
        """
        Destroys the custom chain and all its rules.
        """
        with log.step("Deleting iptables rules"):
            self._run(FLUSH_CUSTOM_CHAIN)
            self._run(DELETE_FORWARD_TO_CUSTOM_CHAIN)
            self._run(DELETE_CUSTOM_CHAIN)

    def ins_rule(self, source, target, action):
        """
        Insert a rule at the beginning of the custom chain.
        """
        action = RuleAction.from_cfg(action)
        rule = Rule(source, target, action)
        # iptables result empty -> rule exists
        rule_exists = not self._run(rule.check_cmd())
        if rule_exists:
            log.info("Rule", repr(rule), "already exists")
        else:
            self._run(rule.insert_cmd())
            log.info("Inserted new rule:", repr(rule))

    def add_rule(self, source, target, action):
        """
        Adds a rule to the custom chain.
        """
        action = RuleAction.from_cfg(action)
        rule = Rule(source, target, action)
        # iptables result empty -> rule exists
        rule_exists = not self._run(rule.check_cmd())
        if rule_exists:
            log.info("Rule", repr(rule), "already exists")
        else:
            self._run(rule.append_cmd())
            log.info("Appended new rule:", repr(rule))

    def del_rule(self, source, target, action):
        """
        Deletes the rule from the custom chain.
        """
        action = RuleAction.from_cfg(action)
        rule = Rule(source, target, action)
        rule_exists = not self._run(rule.check_cmd())
        if rule_exists:    # iptables result empty -> rule exists
            self._run(rule.delete_cmd())
            log.info("Deleted rule:", repr(rule))
        else:
            log.info("Rule", repr(rule), "not found")

    def _parse_rules(self):
        lines = self._run(LIST_CUSTOM_RULES).splitlines()

        parsed = []
        unparsed = []

        for line in lines[1:]:  # skip chain creation rule
            rule = Rule.parse(line)
            if isinstance(rule, Rule):
                parsed.append(rule)
            else:
                unparsed.append(rule)

        return parsed, unparsed

    def _run(self, cmd):
        return self.gateway.exec_run(cmd)


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
    def parse(line):
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

        source = _parse_network_address(ip_source)
        target = _parse_network_address(ip_target)

        return Rule(source, target, RuleAction(ip_action))

    def __init__(self, source, target, action):
        self.source = source
        self.target = target
        self.action = action

    def insert_cmd(self):
        """
        Returns an insert command for the current rule.
        """
        ip_source = _name_to_ip(self.source)
        ip_target = _name_to_ip(self.target)
        return "iptables -I " + IPTABLES_CHAIN + \
            " -s " + ip_source + \
            " -d " + ip_target + \
            " -j " + self.action.value

    def append_cmd(self):
        """
        Returns a create command for the current rule.
        """
        ip_source = _name_to_ip(self.source)
        ip_target = _name_to_ip(self.target)
        return "iptables -A " + IPTABLES_CHAIN + \
            " -s " + ip_source + \
            " -d " + ip_target + \
            " -j " + self.action.value

    def check_cmd(self):
        """
        Returns a check command for the current rule.
        """
        ip_source = _name_to_ip(self.source)
        ip_target = _name_to_ip(self.target)
        return "iptables -C " + IPTABLES_CHAIN + \
            " -s " + ip_source + \
            " -d " + ip_target + \
            " -j " + self.action.value

    def delete_cmd(self):
        """
        Returns a delete command for the current rule.
        """
        ip_source = _name_to_ip(self.source)
        ip_target = _name_to_ip(self.target)
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


def _parse_network_address(string):
    env = Iptables.env
    network = ip_network(string)
    if network.num_addresses == 1:  # "/32"
        ip = str(network.network_address)
        name = env.network.ip_pool.get_container_name_for_ip(ip)
        if name:
            return name
        return ip

    if network == env.network.address:
        return "*"

    return str(network)


def _name_to_ip(name):
    env = Iptables.env
    if name == "*":
        return str(env.network.address)
    container_ip = env.network.ip_pool.get_ip_for_container_name(name)
    if container_ip:
        return str(container_ip)

    return name
