from enum import Enum

from logging import _levelToName, CRITICAL, ERROR, WARNING, INFO, DEBUG, NOTSET


class Level(Enum):
    CRITICAL = _levelToName[CRITICAL]
    ERROR = _levelToName[ERROR]
    WARNING = _levelToName[WARNING]
    INFO = _levelToName[INFO]
    DEBUG = _levelToName[DEBUG]
    NOTSET = _levelToName[NOTSET]
