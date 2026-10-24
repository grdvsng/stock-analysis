from .base import Base
from sqlalchemy import String
from ....types.time import DurationShortType, DurationValue
from sqlalchemy.orm import Mapped, Relationship, mapped_column


class Config(Base):
    logging = Relationship("Logging", uselist=False)
    update_interval_type: Mapped[DurationShortType] = mapped_column(String(10), default="minutes")
    update_interval_value: Mapped[DurationValue] = mapped_column()
