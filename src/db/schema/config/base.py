from ..base import Base as B
from sqlalchemy.orm import Mapped, mapped_column


class Base(B):
    __abstract__ = True
    active: Mapped[bool] = mapped_column()
