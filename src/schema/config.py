from .base import Base
from pathlib import Path
from datetime import datetime
from ..types.time import DurationShortType, DurationLongType, DurationValue
from ..log.level import Level
from ..log.formatter import Format


class Logging(Base):
    level: Level = Level.INFO
    rotation_type: DurationLongType = "days"
    rotation_value: DurationValue = 5.0
    folder: Path = Path("/var/log/stock-analysis")
    format: Format = Format.AVRO


class Config(Base):
    host: str = "localhost"
    port: int = 4004
    update_interval_type: DurationShortType = "minutes"
    update_interval_value: DurationValue = 5.0
    logging: Logging = Logging()

    created_at: datetime = datetime.now()
    updated_at: datetime = datetime.now()
