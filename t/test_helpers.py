from .base import TestCase
from src.helpers.string import camel_to_snake


class TestString(TestCase):
    def test_camel_to_snake(self) -> None:
        self.assertEqual(camel_to_snake("ContractInfo"), "contract_info")
        self.assertEqual(camel_to_snake("Config"), "config")
        self.assertEqual(camel_to_snake("HTTPServer"), "http_server")
