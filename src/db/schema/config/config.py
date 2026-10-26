from .base import Base
from src.types.time import DurationShortType, DurationValue
from sqlalchemy import String, Integer
from sqlalchemy.orm import Mapped, Relationship, mapped_column, backref
from sqlalchemy import ForeignKey
from ..validation import validate_host, validate_port


@validate_host("host")
@validate_port("port")
class Config(Base):
    host: Mapped[str] = mapped_column(String(255))
    port: Mapped[int] = mapped_column(Integer)
    update_interval_type: Mapped[DurationShortType] = mapped_column(default="minutes")
    update_interval_value: Mapped[DurationValue] = mapped_column()

    logging: Mapped["Logging"] = Relationship("Logging", backref=backref("config", uselist=False))
    logging_id: Mapped[int] = mapped_column(ForeignKey("logging.id"))
