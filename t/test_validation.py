from .base import TestCase
from src.suppliers.constants import ALLOWED_SUPPLIER_TYPES, SupplierType
from src.db.schema.validation import (
    ValidationError,
    validate_choice,
    validate_host,
    validate_number,
    validate_port,
)


class TestValidation(TestCase):
    def test_number(self) -> None:
        validator = validate_number(1, 10)

        self.assertEqual(validator("value", 5), 5)

        with self.assertRaises(ValidationError):
            validator("value", 11)

    def test_port(self) -> None:
        validator = validate_port()

        self.assertEqual(validator("port", 4004), 4004)

        for port in (0, 65536):
            with self.assertRaises(ValidationError):
                validator("port", port)

    def test_choice(self) -> None:
        validator = validate_choice(ALLOWED_SUPPLIER_TYPES)

        self.assertEqual(validator("type", SupplierType.IBK.value), SupplierType.IBK.value)

        with self.assertRaises(ValidationError):
            validator("type", "unknown")

    def test_host_rejects_invalid(self) -> None:
        validator = validate_host(None)

        for host in ("", "localhost:4004", "http://localhost"):
            with self.assertRaises(ValidationError):
                validator("host", host)

    def test_host_accepts_plain_host(self) -> None:
        validator = validate_host(None)

        for host in ("localhost", "example.com", "127.0.0.1"):
            self.assertEqual(validator("host", host), host)
