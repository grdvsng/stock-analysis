from .base import Base
from datetime import datetime


class Metric(Base):
    price: float
    size: float
    bid: float
    ask: float
    bid_size: float
    ask_size: float

    created_at: datetime = datetime.now()
