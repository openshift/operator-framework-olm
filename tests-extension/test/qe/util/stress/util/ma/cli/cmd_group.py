import click
import logging
import sys
from ma.cli.cmd_check_ccpu import check_ccpu
import ma.helper.util as util
from ma import version
from ma.helper.const import CONTEXT_SETTINGS

logger = logging.getLogger(__name__)


def print_version(ctx, param, value):
    if not value or ctx.resilient_parsing:
        return
    click.echo("ma v{}".format(version()))
    click.echo("python v{}".format(sys.version))
    ctx.exit()


@click.group(context_settings=CONTEXT_SETTINGS)
@click.pass_context
@click.option(
    "-V",
    "--version",
    is_flag=True,
    callback=print_version,
    expose_value=False,
    is_eager=True,
)
@click.option(
    "-v",
    "--debug",
    help="enable debug logging",
    is_flag=True,
    default=False)
def cli(ctx, debug):
    util.init_logging(logging.DEBUG if debug else logging.INFO)
    is_help = False
    for k in CONTEXT_SETTINGS["help_option_names"]:
        if k in sys.argv:
            is_help = True
            break
    if not is_help:
        logger.info("start to handle sub command")
        pass

cli.add_command(check_ccpu)
