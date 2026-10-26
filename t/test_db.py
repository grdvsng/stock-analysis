from .base import TestCase
from src.db.schema import Meta


class TestDatabase(TestCase):
    async def test_create_engine_creates_tables(self) -> None:
        await self.engine()

        connection = await self.connect(self.name)

        try:
            rows = await connection.fetch(
                "SELECT tablename FROM pg_tables WHERE schemaname = 'public'"
            )
        finally:
            await connection.close()

        self.assertEqual({row["tablename"] for row in rows}, set(Meta.tables))

    async def test_create_engine_is_idempotent(self) -> None:
        await self.engine()
        await self.engine()

    async def test_init_stores_default_config(self) -> None:
        await self.engine()

        connection = await self.connect(self.name)

        try:
            configs = await connection.fetchval("SELECT count(*) FROM config WHERE active")
            loggings = await connection.fetchval("SELECT count(*) FROM logging")
        finally:
            await connection.close()

        self.assertEqual(configs, 1)
        self.assertEqual(loggings, 1)
