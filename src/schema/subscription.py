from .base import Base
from .contract import Contract
from datetime import datetime
from .indicator import Indicator


class Subscription(Base):
    name: str
    description: str = ""

    contract: Contract
    indicators: list[Indicator] = []

    created_at: datetime = datetime.now()
    updated_at: datetime = datetime.now()
