import inifile

from typing import Awaitable, List
from pathlib import Path
from .schema import Meta, Config, Logging
from sqlalchemy import select, insert, update
from ..schema.config import Config as DefaultConfig
from sqlalchemy.ext.asyncio import create_async_engine, AsyncEngine, AsyncSession
from sqlalchemy_utils import database_exists, create_database


async def init(engine: AsyncEngine) -> Awaitable[None]:
    async with engine.begin() as conn:
        await conn.run_sync(Meta.create_all)
        await conn.commit()
    async with AsyncSession(engine) as conn:
        configs: List[Config] = (await conn.execute(select(Config).where(Config.active == True).order_by(Config.updated_at))).scalars().all()

        if len(configs) == 0:
            default_config = DefaultConfig().model_dump()
            logging_id = (await conn.execute(insert(Logging).values(**default_config.pop("logging")).returning(Logging.id))).scalar()
            await conn.execute(insert(Config).values( **default_config,logging_id=logging_id))
            await conn.commit()

            configs = (await conn.execute(select(Config).where(Config.active == True).order_by(Config.updated_at))).scalars().all()
            config = configs[0]
        elif len(configs) == 1:
            config = configs[0]
        else:
            config = configs[0]
            await conn.execute(update(Config).where(Config.id != config.id).values(active=False))
            await conn.commit()


async def create_engine(config_path: Path=Path(__file__).parent / "config.ini", echo: bool = True) -> Awaitable[AsyncEngine]:
    config = inifile.IniFile(config_path)
    database = config.section_as_dict("database")
    connection = config.section_as_dict("connection")

    user = database["user"]
    password = database["password"]
    host = connection["host"]
    port = connection["port"]
    dbname = connection["dbname"]
    url_template = "postgresql{subprotocl}://{user}:{password}@{host}:{port}/{dbname}"
    sync_url = url_template.format(subprotocl="", user=user, password=password, host=host, port=port, dbname=dbname)
    async_url = url_template.format(subprotocl="+asyncpg", user=user, password=password, host=host, port=port, dbname=dbname)

    if not database_exists(sync_url):
        create_database(sync_url)

    engine = create_async_engine(async_url, echo=echo)
    await init(engine)

    return engine
