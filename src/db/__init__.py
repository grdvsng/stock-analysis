import inifile
import asyncio


from typing import Awaitable
from pathlib import Path
from .schema import meta, Config
from sqlalchemy.ext.asyncio import create_async_engine, AsyncEngine


async def init(engine: AsyncEngine) -> Awaitable[None]:
    async with engine.begin() as conn:
        await conn.run_sync(meta.create_all)
    
    print(Config)
    
    
async def create_engine(config_path: Path=Path(__file__).parent / "config.ini") -> Awaitable[AsyncEngine]:
    config = inifile.IniFile(config_path)
    user = config["database"]["user"]
    password = config["database"]["password"]
    host = config["connection"]["host"]
    port = config["connection"]["port"]
    dbname = config["connection"]["dbname"]
    timeout  = config["connection"]["timeout"]
    
    engine = create_async_engine(f"postgresql+asyncpg://{user}:{password}@{host}:{port}/{dbname}?connect_timeout={timeout}", echo=True)
    
    await init(engine)
    
    return engine
