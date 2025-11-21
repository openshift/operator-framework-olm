import logging
import numpy as np

logger = logging.getLogger(__name__)

def z_score(timestamps, values, threshold):
    # (default threshold = 3)
    preliminary_anomalies = []
    mean = np.mean(values)
    std = np.std(values)
    logger.info(f"mean: {mean}, std: {std}")
    for i, value in enumerate(values):
        if std == 0:
            preliminary_anomalies.append((timestamps[i], value))
            continue
        if value > mean + threshold * std:
            preliminary_anomalies.append((timestamps[i], value))
    return preliminary_anomalies


def moving_window_statistics (preliminary_anomalies, timestamps, values, window_size, window_threshold):
    refined_anomalies = []
    # Refinement with moving window (default window_size = 30 points, default threshold = 3)
    logger.info(f"window_size: {window_size}, window_threshold: {window_threshold}")
    for ts, value in preliminary_anomalies:
        index = timestamps.index(ts)
        if index < window_size:
            continue
        window = values[index - window_size : index]
        window_mean = np.mean(window)
        window_std = np.std(window)
        if window_std == 0:
            refined_anomalies.append((ts, value))
            continue
        if value > window_mean + window_threshold * window_std:
            refined_anomalies.append((ts, value))
    return refined_anomalies

def watermark(refined_anomalies, watermark):
    anomalies = []
    for ts, value in refined_anomalies:
        if value > watermark:
            anomalies.append((ts, value))
    return anomalies
