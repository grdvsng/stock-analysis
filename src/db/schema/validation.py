import re

from ...types.numeric import Number
from ...types.result import Result
from urllib.parse import urlparse
from sqlalchemy.orm import validates, DeclarativeBase
from typing import Callable, Optional, Type, TypeVar, Union, Iterable, Hashable


M = TypeVar("M", bound="DeclarativeBase")
T = TypeVar("T")

HOST_REGEX = re.compile(
    r"^(?:(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}|localhost|(?:\d{1,3}\.){3}\d{1,3})$"
)

class ValidationError(ValueError):
    def __init__(self, key: str, message: str):
        self.key = key
        self.message = message
        super().__init__(f"{key}: {message}")


def validate_host(key: Optional[str]) -> Callable[[Type[M]], Type[M]]:
    def _(key: str, value: Optional[str] = None) -> str:
        if not value:
            raise ValidationError(key, "Host cannot be empty")
        try:
            parsed = urlparse(value)

            if not HOST_REGEX.match(value):
                raise ValidationError(key, "Host does not match required pattern")

            if parsed.port is not None:
                raise ValidationError(key, "Port should not be specified in host")

            if parsed.scheme:
                raise ValidationError(key, "Scheme should not be specified in host")
        except Exception as error:
            raise ValidationError(key, f"Invalid host: {error}")

        return value

    if key is not None:
        return validate(key, _)

    return _


def validate_choice(
    values: Iterable[Hashable],
    key: Optional[str] = None,
) -> Union[
    Callable[[str, Optional[Number]], Result[Number, ValidationError]],
    Callable[[Type[M]], Type[M]],
]:
    allowed_values = set(values)

    def _(key: str, value: Optional[Number] = None) -> Number:
        if not value in allowed_values:
            raise ValidationError(key, f"Value must be one of {allowed_values}")

        return value

    if key is not None:
        return validate(key, _)

    return _


def validate_number(min: Number, max: Number, key: Optional[str] = None) -> Union[
    Callable[[str, Optional[Number]], Result[Number, ValidationError]],
    Callable[[Type[M]], Type[M]],
]:
    def _(key: str, value: Optional[Number] = None) -> Result[Number, ValidationError]:
        if not (min <= value <= max):
            raise ValidationError(key, f"Value must be between {min} and {max}")
        return value

    if key is not None:
        return validate(key, _)

    return _


def validate_port(key: Optional[str] = None) -> Union[
    Callable[[str, Optional[Number]], Result[Number, ValidationError]],
    Callable[[Type[M]], Type[M]],
]:
    return validate_number(1, 65535, key=key)


def validate(
    key: str, validator_type: Callable[[str, Optional[T]], Result[T, ValidationError]]
) -> Callable[[Type[M]], Type[M]]:
    def wrapper(cls: Type[M]) -> Type[M]:
        def v(
            _: M, key: str, value: Optional[T]
        ) -> Result[T, ValidationError]:
            return validator_type(key, value)

        validates(key)(v)

        return cls

    return wrapper
