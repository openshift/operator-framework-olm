import logging
import json
import os
import matplotlib.pyplot as plt
from ma.helper.const import *
from pathlib import Path



def init_logging(log_level=logging.INFO):
    logging.basicConfig(
        # format="%(module)s: %(asctime)s: %(levelname)s: %(message)s",
        format="%(asctime)s: %(levelname)s: %(message)s",
        datefmt="%Y-%m-%dT%H:%M:%SZ",
        level=log_level,
    )

    loggers = logging.Logger.manager.loggerDict
    for k in loggers.keys():
        if "requests" in k or "urllib3" in k or "gssapi" in k:
            logger = logging.getLogger(k)
            logger.setLevel(logging.WARNING)
        if "requests_kerberos" in k:
            logger = logging.getLogger(k)
            logger.setLevel(logging.CRITICAL)

def draw_figure(timestamps_format, values, odir, base_name_wo_ext):
    plt.figure(figsize=(FIGURE_WIDTH, FIGURE_HEIGHT))
    plt.plot(timestamps_format, values, marker='o')
    # plt.xticks(rotation=45)
    plt.ylabel('CPU Usage')
    plt.title('CPU Usage Over Time')
    plt.tight_layout()
    saved_file_wo_ext = os.path.join(odir, base_name_wo_ext)
    plt.savefig(saved_file_wo_ext+"_figure.pdf")
    plt.close()

def convert_screening(anomalies, file):
    formatted_data = [
        {"timestamp": ts, "value": round(val, 15)}
        for ts, val in anomalies
    ]

    output_path = Path(file)
    output_path.write_text(
        json.dumps(formatted_data, indent=2, ensure_ascii=False)
    )
