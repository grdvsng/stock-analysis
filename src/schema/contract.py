from .base import Base
from datetime import datetime
from typing import List, Optional
from ..suppliers.constants import SupplierType
from ..types.time import DurationLongType, DurationValue, DurationShortType


class Supplier(Base):
    type: SupplierType
    host: str
    port: int
    timeout: int = 30


class Info(Base):
    name: str
    exchange: str
    currency: str
    industry: str
    category: str
    subcategory: str
    primary_exchange: str
    country: str
    trading_hours: List[datetime]
    liquid_hours: List[datetime]


class Contract(Base):
    id: Optional[int] = None

    info: Info
    supplier: Supplier

    active: bool = True
    ttl_type: DurationLongType = "days"
    ttl_value: DurationValue = 60
    tick_interval_type: DurationShortType = "seconds"
    tick_interval_value: DurationValue = 0.5
    reload_info_interval_type: DurationLongType = "days"
    reload_info_interval_value: DurationValue = 1

    created_at: datetime = datetime.now()
    updated_at: datetime = datetime.now()
