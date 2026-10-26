from .base import Base
from sqlalchemy.orm import Mapped, mapped_column, Relationship, backref
from sqlalchemy import ForeignKey
from src.types.time import DurationLongType, DurationValue, DurationShortType


class Contract(Base):
    ttl_type: Mapped[DurationLongType] = mapped_column(default="days")
    ttl_value: Mapped[DurationValue] = mapped_column(default=60)
    tick_interval_type: Mapped[DurationShortType] = mapped_column(default="seconds")
    tick_interval_value: Mapped[DurationValue] = mapped_column(default=0.5)
    reload_info_interval_type: Mapped[DurationLongType] = mapped_column(default="days")
    reload_info_interval_value: Mapped[DurationValue] = mapped_column(default=1)

    supplier = Relationship("Supplier", backref=backref("contract", uselist=False))
    supplier_id: Mapped[int] = mapped_column(ForeignKey("supplier.id"))
