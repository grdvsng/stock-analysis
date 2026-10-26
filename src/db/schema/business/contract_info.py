from datetime import datetime
from typing import List
from sqlalchemy import String, ForeignKey
from ..base import Base
from sqlalchemy.orm import Mapped, Relationship, mapped_column, backref


class ContractInfo(Base):
    id: Mapped[str] = mapped_column(String(36), primary_key=True)
    name: Mapped[str] = mapped_column(String(255))
    exchange: Mapped[str] = mapped_column(String(255))
    currency: Mapped[str] = mapped_column(String(255))
    industry: Mapped[str] = mapped_column(String(255))
    category: Mapped[str] = mapped_column(String(255))
    subcategory: Mapped[str] = mapped_column(String(255))
    primary_exchange: Mapped[str] = mapped_column(String(255))
    country: Mapped[str] = mapped_column(String(255))
    trading_hours: Mapped[List[datetime]] = mapped_column()
    liquid_hours: Mapped[List[datetime]] = mapped_column()

    contract = Relationship("Contract", backref=backref("contract_info", uselist=False))
    contract_id: Mapped[int] = mapped_column(ForeignKey("contract.id"))
