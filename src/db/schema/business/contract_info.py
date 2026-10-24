from datetime import datetime
from typing import List
from sqlalchemy import String
from ..base import Base
from sqlalchemy.orm import Mapped, Relationship, mapped_column


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

    metrics = Relationship("Metric", back_populates="contract_info")
    contract = Relationship("Contract", back_populates="contract_info", uselist=False)

