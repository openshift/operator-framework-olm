import click
import logging
from ma.helper.containercpu import ContainerCPU

logger = logging.getLogger(__name__)


@click.command()
@click.pass_context
@click.option(
    "-i",
    "--metrics_result_file",
    type=click.Path(
        exists=True,
        file_okay=True,
        dir_okay=False,
        readable=True,
        resolve_path=True
    ),
    required=True,
    help="the result file of the metrics"
)
@click.option(
    "-o",
    "--output_dir",
    default="./",
    type=click.Path(
        exists=True,
        file_okay=False,
        dir_okay=True,
        readable=True,
        resolve_path=True
    ),
    help="the directory of generated metrics figure"
)
@click.option(
    "--zscore_threshold",
    type=int,
    default=3,
    required=False,
    help="the threshold for z-score"
)
@click.option(
    "--window_size",
    type=int,
    default=18,
    required=False,
    help="the size of moving window"
)
@click.option(
    "--window_threshold",
    type=int,
    default=3,
    required=False,
    help="the threshold of moving window"
)
@click.option(
    "--watermark",
    type=int,
    default=20,
    required=False,
    help="the abnoarm cpu usage calucalted by 100"
)
@click.option(
    "--anomalies_threshold",
    type=int,
    default=2,
    required=False,
    help="the anomalies threshold to determine if the checking fails or not"
)
def check_ccpu(ctx,
    metrics_result_file,
    output_dir,
    zscore_threshold,
    window_size,
    window_threshold,
    watermark,
    anomalies_threshold):
    """
    Check if cpu usage is expected
    """
    try:
        ccpu = ContainerCPU(metrics_result_file,
            output_dir,
            zscore_threshold,
            window_size,
            window_threshold,
            watermark,
            anomalies_threshold)
        ccpu.handle()
        ccpu.ok_or_not()
        # print(ccpu.get_preliminary_screening())
        # print(ccpu.get_refinement_screening())
        # print(ccpu.get_final_screening())
    except Exception as e:
        logger.exception("checking container cpu failing")
        raise

