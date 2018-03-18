"""
Utility functions for logging the progress.
"""
import contextlib
import sys
from termcolor import colored


@contextlib.contextmanager
def step(*args):
    """
    Log a long-running step in progress.
    """
    print(*args, end="")
    print("...", end="")
    sys.stdout.flush()
    yield
    print(colored("OK", "green"))
    sys.stdout.flush()


def section(*args):
    """
    Write a log entry for a larger section.
    """
    print()
    bold_args = [colored(arg, attrs=['bold']) for arg in args]
    print(colored("==> ", attrs=['bold']), end="")
    print(*bold_args, end="")
    print()


def info(*args):
    """
    Log an information message.
    """
    print(*args)


def fatal(*args):
    """
    Log a fatal error and exit the app.
    """
    red_args = [colored(arg, 'red') for arg in args]
    print(*red_args, end="")
    print()
    sys.exit(-1)
