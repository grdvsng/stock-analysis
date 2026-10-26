import json

from fastavro import reader as avro_reader
from .base import TestCase
from src.log.level import Level
from src.log.handler import Tsv as TsvHandler, Avro as AvroHandler


class TestLog(TestCase):
    def test_json_handler_writes_record(self) -> None:
        self.log.warning("hello %s", "world")
        self.handler.flush()

        lines = (self.folder / "test.log").read_text().splitlines()
        record = json.loads(lines[-1])

        self.assertEqual(record["level"], Level.WARNING.value)
        self.assertEqual(record["message"], "hello world")
        self.assertEqual(record["func"], "test_json_handler_writes_record")

    def test_tsv_handler_writes_record(self) -> None:
        path = self.folder / "tsv.log"
        handler = TsvHandler(path)
        self.log.addHandler(handler)

        try:
            self.log.error("tsv message")
        finally:
            self.log.removeHandler(handler)
            handler.close()

        fields = path.read_text().splitlines()[-1].split("\t")

        self.assertEqual(fields[3], Level.ERROR.value)
        self.assertEqual(fields[-1], "tsv message")

    def test_avro_handler_writes_record(self) -> None:
        path = self.folder / "test.avro"
        handler = AvroHandler(Level.INFO, path)
        self.log.addHandler(handler)

        try:
            self.log.debug("filtered")
            self.log.warning("avro %s", "message")
        finally:
            self.log.removeHandler(handler)
            handler.close()

        with path.open("rb") as stream:
            records = list(avro_reader(stream))

        self.assertEqual(len(records), 1)
        self.assertEqual(records[0]["level"], Level.WARNING.value)
        self.assertEqual(records[0]["message"], "avro message")
        self.assertEqual(records[0]["func"], "test_avro_handler_writes_record")
