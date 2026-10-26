import enum

from pathlib import Path
from datetime import datetime
from typing import Any, List, Literal, Optional
from sqlalchemy import ARRAY, DateTime, Dialect, Enum, String
from sqlalchemy.types import TypeDecorator, TypeEngine


class PathType(TypeDecorator[Path]):
    impl = String(1024)
    cache_ok = True

    def process_bind_param(self, value: Optional[Path], _: Dialect) -> Optional[str]:
        if value is None:
            return None

        return str(value)

    def process_result_value(self, value: Optional[str], _: Dialect) -> Optional[Path]:
        if value is None:
            return None

        return Path(value)


def enum_values(enum_class: type[enum.Enum]) -> list[str]:
    return [str(member.value) for member in enum_class]


TIMESTAMP = DateTime(timezone=True)

TYPE_ANNOTATION_MAP: dict[Any, TypeEngine[Any]] = {
    datetime: TIMESTAMP,
    Path: PathType(),
    List[datetime]: ARRAY(TIMESTAMP),
    list[datetime]: ARRAY(TIMESTAMP),
    Literal: Enum(native_enum=False, create_constraint=True),
    enum.Enum: Enum(
        enum.Enum,
        native_enum=False,
        create_constraint=True,
        values_callable=enum_values,
    ),
}
