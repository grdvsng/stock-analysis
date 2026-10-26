from .base import Base
from ..types.time import DurationType
from datetime import datetime
from typing import Optional


class Indicator(Base):
    id: Optional[int] = None
    name: str
    period_type: DurationType = "minutes"
    period_value: int = 1
    formula: str
    description: str = ""

    created_at: datetime = datetime.now()
    updated_at: datetime = datetime.now()
