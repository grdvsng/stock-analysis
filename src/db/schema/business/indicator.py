from ..base import Base
from sqlalchemy import String
from sqlalchemy.orm import Mapped, mapped_column
from src.types.time import DurationType


class Indicator(Base):
    name: Mapped[str] = mapped_column(String(255))
    period_type: Mapped[DurationType] = mapped_column(default="minutes")
    period_value: Mapped[int] = mapped_column(default=1)
    formula: Mapped[str] = mapped_column(String())
    description: Mapped[str] = mapped_column(String(1000))
