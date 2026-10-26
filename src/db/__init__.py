import inifile

from typing import Awaitable
from pathlib import Path
from .schema import Meta, Config
from sqlalchemy import select
from sqlalchemy.ext.asyncio import create_async_engine, AsyncEngine, AsyncSession


async def init(engine: AsyncEngine) -> Awaitable[None]:
    async with engine.begin() as conn:
        await conn.run_sync(Meta.create_all)

        result = await conn.execute(select(Config).where(Config.active == True))
        config = result.scalars().all()
        print(config)



async def create_engine(config_path: Path=Path(__file__).parent / "config.ini") -> Awaitable[AsyncEngine]:
    config = inifile.IniFile(config_path)
    database = config.section_as_dict("database")
    connection = config.section_as_dict("connection")

    user = database["user"]
    password = database["password"]
    host = connection["host"]
    port = connection["port"]
    dbname = connection["dbname"]
    timeout  = connection["timeout"]

    engine = create_async_engine(f"postgresql+asyncpg://{user}:{password}@{host}:{port}/{dbname}?command_timeout={timeout}", echo=True)

    await init(engine)

    return engine
