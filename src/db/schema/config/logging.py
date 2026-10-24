from .base import Base
from pathlib import Path
from ....src.types.time import DurationLongType, DurationValue
from sqlalchemy import String
from ....src.log.formater import Format
from ....src.log.level import Level
from sqlalchemy.orm import Mapped, mapped_column, Relationship


class Logging(Base):
    config = Relationship("Config", uselist=False) 
    level: Mapped[Level] = mapped_column(String(10), default="INFO")
    rotation_type: Mapped[DurationLongType] = mapped_column(String(10), default="days")
    rotation_value: Mapped[DurationValue] = mapped_column()
    folder: Mapped[Path] = mapped_column(String(255))
    format: Mapped[Format] = mapped_column(String(10), default="DEFAULT")
