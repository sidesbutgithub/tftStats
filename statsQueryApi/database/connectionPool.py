from psycopg2 import pool
import os

class connectionPool():
    def __init__(self):
        self.pool = pool.SimpleConnectionPool((os.getenv("CONNECTION_POOL_MIN") if os.getenv("CONNECTION_POOL_MIN") else 3),
                                                  (os.getenv("CONNECTION_POOL_MAX") if os.getenv("CONNECTION_POOL_MAX") else 10),
                                                  os.getenv("PG_URI")
                                                  )

    def getConnection(self):
        return self.pool.getconn()
    