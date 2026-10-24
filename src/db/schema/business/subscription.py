from ..base import Base
from sqlalchemy import String
from sqlalchemy.orm import Mapped, mapped_column, Relationship


class Subscription(Base):
    name: Mapped[str] = mapped_column(String(255))
    description: Mapped[str] = mapped_column(String(1000))
    contract = Relationship("Contract", uselist=False)
    indicators = Relationship("Indicator")