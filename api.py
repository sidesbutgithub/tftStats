from fastapi import FastAPI, Query
from fastapi import Depends
from psycopg2 import pool
from psycopg2 import extensions
import os
from dotenv import load_dotenv
from contextlib import asynccontextmanager
from typing import Annotated


load_dotenv()

connPool = pool.SimpleConnectionPool(3, 10, 
                                    database = os.getenv("PG_DATABASE"), 
                                    user = os.getenv("PG_USERNAME"), 
                                    host= os.getenv("PG_HOST"),
                                    password = os.getenv("PG_PASSWORD"),
                                    port = os.getenv("PG_PORT"))

def getConnection():
    return connPool.getconn()

@asynccontextmanager
async def lifespan(app: FastAPI):
    app.conn = getConnection()
    yield
    connPool.putconn(app.conn)


app = FastAPI(lifespan=lifespan)

def getDB():
    db = app.conn
    cursor = db.cursor()
    try:
        yield cursor
    finally:
        cursor.close()

@app.get("/api")
async def get_unit(unitName: str = "", starLevel: int = 0, items: Annotated[list[str], Query()] = [], db: extensions.cursor = Depends(getDB),):
    if unitName != "":
        if starLevel != 0:
            db.execute('''
                        SELECT AVG(placement)
                        FROM units
                        WHERE (unitname = %s
                        AND starlevel = %s
                        AND items @> %s::varchar[])
                       ''', (unitName, starLevel, items,))
            return {"AVP":db.fetchone()}
        else:
            db.execute('''
                        SELECT AVG(placement)
                        FROM units
                        WHERE (unitname = %s
                        AND items @> %s::varchar[])
                       ''', (unitName, items,))
            return {"AVP":db.fetchone()}

    db.execute('''
               SELECT AVG(placement)
               FROM units
               WHERE items @> %s::varchar[])
               ''', (items,))
    return {"AVP":db.fetchone()}
