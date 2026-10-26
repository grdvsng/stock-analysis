from ...helpers.string import camel_to_snake
from getpass import getuser
from datetime import datetime, UTC
from sqlalchemy.orm import declared_attr
from sqlalchemy import String, BigInteger
from sqlalchemy.ext.asyncio import AsyncAttrs
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column

from sqlalchemy import MetaData
from .type_map import TYPE_ANNOTATION_MAP

meta = MetaData()


def utc_now() -> datetime:
    return datetime.now(UTC)


class Base(AsyncAttrs, DeclarativeBase):
    __abstract__ = True
    __table_args__ = {"extend_existing": True}

    metadata = meta
    type_annotation_map = TYPE_ANNOTATION_MAP

    id: Mapped[int] = mapped_column(BigInteger, primary_key=True, autoincrement=True)
    created_at: Mapped[datetime] = mapped_column(
        default=utc_now, sort_order=1, index=True
    )
    updated_at: Mapped[datetime] = mapped_column(
        default=utc_now, onupdate=utc_now
    )
    created_by: Mapped[str] = mapped_column(
        String(64), sort_order=1, default=getuser
    )

    @declared_attr.directive
    def __tablename__(cls) -> str:
        return camel_to_snake(cls.__name__)
