import uuid

from getpass import getuser
from datetime import datetime
from sqlalchemy import DateTime, String
from sqlalchemy.ext.asyncio import AsyncAttrs
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column

from sqlalchemy import MetaData

meta = MetaData()


class Base(AsyncAttrs, DeclarativeBase):
    __abstract__ = True
    __table_args__ = {"extend_existing": True, "metadata": meta}

    id: Mapped[uuid.UUID] = mapped_column(primary_key=True, default=uuid.uuid4)
    created_at: Mapped[DateTime] = mapped_column(
        default_factory=datetime.now, sort_order=1
    )
    updated_at: Mapped[DateTime] = mapped_column(
        default_factory=datetime.now, onupdate=datetime.now
    )
    created_by: Mapped[str] = mapped_column(
        String(64), sort_order=1, default_factory=getuser
    )
