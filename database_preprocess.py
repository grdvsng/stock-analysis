import asyncio
from argparse import ArgumentParser
from pathlib import Path
from src.db import create_engine

if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument("config_path", type=Path)
    args = parser.parse_args()

    asyncio.run(create_engine(config_path=args.config_path))
