from .base import Base
from datetime import datetime
from .subscription import Subscription
from .indicator import Indicator


class Post(Base):
    indicator: Indicator
    subscription: Subscription

    created_at: datetime = datetime.now()
