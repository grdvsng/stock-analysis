from .base import Base
from ..constants import MAX_TIMEOUT_SECONDS
from sqlalchemy.orm import Mapped, mapped_column
from src.suppliers.constants import ALLOWED_SUPPLIER_TYPES, SupplierType
from ..validation import validate_choice, validate_number, validate_host, validate_port


@validate_host("host")
@validate_port("port")
@validate_number(key="timeout", min=0.001, max=MAX_TIMEOUT_SECONDS)
@validate_choice(key="type", values=ALLOWED_SUPPLIER_TYPES)
class Supplier(Base):
    type: Mapped[SupplierType] = mapped_column()
    host: Mapped[str] = mapped_column()
    port: Mapped[int] = mapped_column()
    timeout: Mapped[int] = mapped_column()
