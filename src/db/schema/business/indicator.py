from ..base import Base
from sqlalchemy import String, ForeignKey
from sqlalchemy.orm import Mapped, mapped_column, Relationship, backref
from ....types.time import DurationType


class Indicator(Base):
    name: Mapped[str] = mapped_column(String(255))
    period_type: Mapped[DurationType] = mapped_column(default="minutes")
    period_value: Mapped[int] = mapped_column(default=1)
    formula: Mapped[str] = mapped_column(String())
    description: Mapped[str] = mapped_column(String(1000))

    contract: Mapped["Contract"] = Relationship("Contract", backref=backref("metrics", uselist=True))
    contract_id: Mapped[int] = mapped_column(ForeignKey("contract.id"))
