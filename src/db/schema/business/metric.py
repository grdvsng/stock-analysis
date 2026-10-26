from ..base import Base
from sqlalchemy.orm import Mapped, mapped_column, Relationship, backref
from sqlalchemy import ForeignKey


class Metric(Base):
    price: Mapped[float] = mapped_column()
    size: Mapped[float] = mapped_column()
    bid: Mapped[float] = mapped_column()
    ask: Mapped[float] = mapped_column()
    bid_size: Mapped[float] = mapped_column()
    ask_size: Mapped[float] = mapped_column()

    contract_id: Mapped[int] = mapped_column(ForeignKey("contract.id"))
