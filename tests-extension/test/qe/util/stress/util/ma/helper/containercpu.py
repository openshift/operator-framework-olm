import json
import os
import logging
from datetime import datetime
from pathlib import Path
import ma.helper.util as util
import ma.helper.algo as algo
from ma.helper.exceptions import ContainerCPUException

logger = logging.getLogger(__name__)


class ContainerCPU:
    """
    Container CPU object is used to check if the Container CPU usage is expected.

    """

    def __init__(self, metrics_result_file, output_dir, zscore_threshold, window_size, window_threshold, watermark, anomalies_threshold):
        self.mrf = metrics_result_file
        self.odir = output_dir
        self.zscore_threshold = zscore_threshold
        self.window_size = window_size
        self.window_threshold = window_threshold
        self.watermark = watermark
        self.anomalies_threshold = anomalies_threshold
        self.preliminary_anomalies = []
        self.refined_anomalies = []
        self.final_anomalies = []

        try:
            self.base_name = os.path.basename(self.mrf)
            self.base_name_wo_ext, self.base_name_ext = os.path.splitext(self.base_name)
        except BaseException as re:
            raise ContainerCPUException("parse the file path failed") from re

        try:
            with open(self.mrf) as f:
                self.data = json.load(f)
            self.values =     [d['value'] for d in self.data]
            self.timestamps = [d['timestamp'] for d in self.data]
            self.timestamps_format = [datetime.strptime(d['timestamp'], "%Y-%m-%dT%H:%M:%S.%fZ") for d in self.data]

        except BaseException as re:
            raise ContainerCPUException("load ccpu metric data failed") from re

    def handle(self):
        self.tendency_chart()
        self.preliminary_screening()
        self.refinement_screening()
        self.final_screening()
        self.convert_preliminary_screening()
        self.convert_refinement_screening()
        self.convert_final_screening()

    def tendency_chart(self):
        """
        it draws the cpu usage figure and save it as pdf
        """
        try:
            util.draw_figure(
                self.timestamps_format,
                self.values,
                self.odir,
                self.base_name_wo_ext
            )
        except BaseException as re:
            raise ContainerCPUException("drawing pdf failed") from re


    def preliminary_screening(self):
        """
        it uses a simple, fast method (e.g., Z-Score) to flag potential anomalies
        """
        try:
            self.preliminary_anomalies = algo.z_score(
                self.timestamps,
                self.values,
                self.zscore_threshold)
        except BaseException as re:
            raise ContainerCPUException("preliminary screening failed") from re

    def refinement_screening(self):
        """
        it applyes stricter, context-aware rules (e.g., moving window statistics) to 
        validate whether the candidates flagged in preliminary_screening are true anomalies
        """
        try:
            self.refined_anomalies = algo.moving_window_statistics(
                self.preliminary_anomalies,
                self.timestamps,
                self.values,
                self.window_size,
                self.window_threshold)
        except BaseException as re:
            raise ContainerCPUException("refinement screening failed") from re

    def final_screening(self):
        try:
            self.final_anomalies = algo.watermark(
                self.refined_anomalies,
                self.watermark)
        except BaseException as re:
            raise ContainerCPUException("final screening failed") from re

    def get_preliminary_screening(self):
        """
        it gets the result of a simple, fast method (e.g., Z-Score) to flag potential anomalies
        """
        return self.preliminary_anomalies

    def get_refinement_screening(self):
        """
        it get results of refinement screening
        """
        return self.refined_anomalies

    def get_final_screening(self):
        """
        it get results of final screening
        """
        return self.final_anomalies

    def convert_preliminary_screening(self):
        """
        it converts the result of a simple, fast method (e.g., Z-Score) to flag potential anomalies
        """
        try:
            util.convert_screening(
                self.preliminary_anomalies,
                os.path.join(self.odir, self.base_name_wo_ext) + "_prescr.json"
            )
        except BaseException as re:
            raise ContainerCPUException("convert preliminary screening failed") from re


    def convert_refinement_screening(self):
        """
        it converts results of refinement screening
        """
        try:
            util.convert_screening(
                self.refined_anomalies,
                os.path.join(self.odir, self.base_name_wo_ext) + "_refscr.json"
            )
        except BaseException as re:
            raise ContainerCPUException("convert refinement screening failed") from re

    def convert_final_screening(self):
        """
        it converts results of final screening
        """
        try:
            util.convert_screening(
                self.final_anomalies,
                os.path.join(self.odir, self.base_name_wo_ext) + "_finscr.json"
            )
        except BaseException as re:
            raise ContainerCPUException("convert final screening failed") from re

    def ok_or_not(self):
        """
        it reports if the result is ok
        """
        try:
            base_path = os.path.join(self.odir, self.base_name_wo_ext)
            result = "pass"
            if len(self.final_anomalies) > self.anomalies_threshold:
                result = "fail"
            output_path = Path(base_path+"_result-"+result)
            output_path.write_text(result)
        except BaseException as re:
            raise ContainerCPUException("check result failed") from re