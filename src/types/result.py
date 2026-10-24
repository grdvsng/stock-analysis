from typing import TypeVar, Union, Never, Generic

T = TypeVar("T")
E = TypeVar("E", bound=Exception)


class TypedError(Never, Generic[E]):
    pass


Result = Union[T, TypedError[E]]
