from ..base import Base
from sqlalchemy import String, ForeignKey
from sqlalchemy.orm import Mapped, mapped_column, Relationship, backref


class Subscription(Base):
    name: Mapped[str] = mapped_column(String(255))
    description: Mapped[str] = mapped_column(String(1000))

    contract = Relationship("Contract", backref=backref("subscription", uselist=False))
    contract_id: Mapped[int] = mapped_column(ForeignKey("contract.id"))

    indicators = Relationship("Indicator", backref=backref("subscription", uselist=True))
    indicator_ids: Mapped[int] = mapped_column(ForeignKey("indicator.id"))
