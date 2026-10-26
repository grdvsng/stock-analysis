from ..base import Base
from sqlalchemy import ForeignKey
from sqlalchemy.orm import Mapped, mapped_column, Relationship, backref


class Post(Base):
    indicator_id: Mapped[int] = mapped_column(ForeignKey("indicator.id"))
