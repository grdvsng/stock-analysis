from enum import Enum

from logging import _levelToName, LOG_LEVELS


class Level(Enum):
    CRITICAL = _levelToName[LOG_LEVELS['CRITICAL']]
    ERROR = _levelToName[LOG_LEVELS['ERROR']]
    WARNING = _levelToName[LOG_LEVELS['WARNING']]
    INFO = _levelToName[LOG_LEVELS['INFO']]
    DEBUG = _levelToName[LOG_LEVELS['DEBUG']]
    NOTSET = _levelToName[LOG_LEVELS['NOTSET']]
