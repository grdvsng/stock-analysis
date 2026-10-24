from .base import Base
from datetime import datetime
from typing import List
from sqlalchemy import String
from sqlalchemy.orm import Mapped, mapped_column, Relationship
from ....src.types.time import DurationLongType, DurationValue, DurationShortType


class Contract(Base):
    ttl: Mapped[DurationLongType] = mapped_column(String(10), default="days")
    minimum_data_type: Mapped[DurationLongType] = mapped_column(
        String(10), default="days"
    )
    minimum_data_value: Mapped[DurationValue] = mapped_column(default=10)
    tick_interval_type: Mapped[DurationShortType] = mapped_column(
        String(10), default="seconds"
    )
    tick_interval_value: Mapped[DurationValue] = mapped_column(default=0.5)
    reload_info_interval_type: Mapped[DurationLongType] = mapped_column(
        String(10), default="days"
    )
    reload_info_interval_value: Mapped[DurationValue] = mapped_column(default=1)

    supplier = Relationship("Supplier", uselist=False)
