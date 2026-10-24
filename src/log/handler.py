from click import Path
from threading import Lock
from src.log.level import Level
from .record import Record
from fastavro import writer as avro_writer, parse_schema
from typing import Dict, Generic, TypeVar, List, Any
from .formater import Json as JsonFormatter, Tsv as TsvFormatter, Avro as AvroFormatter
from logging import FileHandler, Formatter, Handler

F = TypeVar("F", bound="Formatter")


class TextHandler(FileHandler, Generic[F]):
    formatter: F
    terminator = "\n"

    def __init__(self, path: Path) -> None:
        super().__init__(str(path))

    def emit(self, record):
        formatted = self.format.format(record)
        self.stream.write(formatted + self.terminator)


class Json(TextHandler[JsonFormatter]):
    format = JsonFormatter()


class Tsv(TextHandler[TsvFormatter]):
    format = TsvFormatter()


class Avro(Handler):
    formatter = AvroFormatter()
    max_cache_size: int = 1000

    def __init__(self, level: Level, path: Path) -> None:
        self.schema = parse_schema(Record.avro_schema())
        self.cache: List[Dict[str, Any]] = []
        self.path = path
        self.mutex = Lock()

        super().__init__(level=level)

    def flush(self, force: bool = False) -> None:
        with self.mutex:
            if self.cache and (force or len(self.cache) >= self.max_cache_size):
                with open(self.path, "ab") as stream:
                    avro_writer(stream, self.schema, self.cache)
                self.cache.clear()

    def emit(self, record):
        with self.mutex:
            self.cache.append(self.formatter.format(record))

        self.flush()

    def __del__(self):
        self.flush(True)
