import shutil
import logging
import asyncpg

from pathlib import Path
from typing import List
from configparser import ConfigParser
from unittest import IsolatedAsyncioTestCase
from sqlalchemy.ext.asyncio import AsyncEngine
from src.db import create_engine
from src.log.handler import Json as JsonHandler


ROOT = Path(__file__).parent
CONFIG_PATH = ROOT / "config" / "config.ini"
LOGGERS = ("sqlalchemy.engine.Engine",)


class TestCase(IsolatedAsyncioTestCase):
    name: str
    folder: Path
    config_path: Path
    log: logging.Logger
    handler: logging.Handler
    engines: List[AsyncEngine]

    async def asyncSetUp(self) -> None:
        self.name = self._testMethodName.lower()
        self.folder = ROOT / f"tmp_{self.name}"
        self.engines = []

        shutil.rmtree(self.folder, ignore_errors=True)
        self.folder.mkdir(parents=True)

        self.setup_logging()
        self.config_path = self.write_config()

        await self.drop_database()
        await self.admin_execute(f'CREATE DATABASE "{self.name}"')
        self.log.info("database %s created", self.name)

    async def asyncTearDown(self) -> None:
        for engine in self.engines:
            await engine.dispose()

        await self.drop_database()
        self.log.info("database %s dropped", self.name)
        self.teardown_logging()

    def setup_logging(self) -> None:
        self.handler = JsonHandler(self.folder / "test.log")
        self.log = logging.getLogger(f"t.{self.name}")
        self.log.setLevel(logging.DEBUG)

        for logger in (self.log, *map(logging.getLogger, LOGGERS)):
            logger.addHandler(self.handler)

    def teardown_logging(self) -> None:
        for logger in (self.log, *map(logging.getLogger, LOGGERS)):
            logger.removeHandler(self.handler)

        self.handler.close()

    def read_config(self) -> ConfigParser:
        config = ConfigParser(interpolation=None)
        config.read(CONFIG_PATH)

        return config

    def write_config(self) -> Path:
        config = self.read_config()
        config["connection"]["dbname"] = self.name
        path = self.folder / "config.ini"

        with path.open("w") as stream:
            config.write(stream)

        return path

    async def connect(self, dbname: str | None = None) -> asyncpg.Connection:
        config = self.read_config()

        return await asyncpg.connect(
            user=config["database"]["user"],
            password=config["database"]["password"],
            host=config["connection"]["host"],
            port=int(config["connection"]["port"]),
            database=dbname or config["connection"]["dbname"],
        )

    async def admin_execute(self, sql: str) -> None:
        connection = await self.connect()

        try:
            await connection.execute(sql)
        finally:
            await connection.close()

    async def drop_database(self) -> None:
        await self.admin_execute(f'DROP DATABASE IF EXISTS "{self.name}" WITH (FORCE)')

    async def engine(self) -> AsyncEngine:
        engine = await create_engine(self.config_path)
        self.engines.append(engine)

        return engine
