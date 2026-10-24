import asyncio

from src.db import create_engine

if __name__ == "__main__":
    asyncio.run(create_engine())
