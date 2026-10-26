from .base import Base
from pathlib import Path
from src.types.time import DurationLongType, DurationValue
from src.log.formatter import Format
from src.log.level import Level
from sqlalchemy.orm import Mapped, mapped_column


class Logging(Base):
    level: Mapped[Level] = mapped_column(default=Level.INFO)
    rotation_type: Mapped[DurationLongType] = mapped_column(default="days")
    rotation_value: Mapped[DurationValue] = mapped_column()
    folder: Mapped[Path] = mapped_column()
    format: Mapped[Format] = mapped_column(default=Format.JSON)
