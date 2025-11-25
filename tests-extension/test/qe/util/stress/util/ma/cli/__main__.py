from ma.cli.cmd_group import cli


def main():
    try:
        cli(obj={})
    except Exception as e:
        raise SystemExit(e)
