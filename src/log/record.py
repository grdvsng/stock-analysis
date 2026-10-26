import json
import socket

from pathlib import Path
from .level import Level
from datetime import datetime
from getpass import getuser
from pydantic import BaseModel, Field
from pydantic_avro import AvroBase


class Record(BaseModel):
    host: str = Field(default_factory=socket.gethostname)
    user: str = Field(default_factory=getuser)
    timestamp: str = Field(default_factory=lambda: datetime.now().isoformat())
    level: Level
    file: Path
    func: str
    line: int
    message: str

    def __str__(self) -> str:
        return "\t".join(
            [
                self.host,
                self.user,
                self.timestamp,
                self.level.value,
                str(self.file),
                str(self.line),
                self.message,
            ]
        )

    @classmethod
    def avro_schema(cls) -> str:
        model = type(cls.__name__, (AvroBase, cls), {})
        return model.avro_schema()
