from enum import Enum
from .record import Record
from typing import Dict, Any
from logging import Formatter, Record as BasicRecord


class Basic(Formatter):
    def convert(self, record: BasicRecord) -> Record:
        return Record(
            level=record.levelname,
            file=record.pathname,
            func=record.funcName,
            line=record.lineno,
            message=record.getMessage(),
        )


class Json(Basic):
    def format(self, record: BasicRecord) -> str:
        return self.convert(record).model_dump_json()


class Tsv(Basic):
    def format(self, record: BasicRecord) -> str:
        return str(self.convert(record))


class Avro(Basic):
    def format(self, record: BasicRecord) -> Dict[str, Any]:
        return self.convert(record).model_dump()


class Format(Enum):
    JSON = "json"
    TSV = "tsv"
    AVRO = "avro"
