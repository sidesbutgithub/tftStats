from fastapi import FastAPI
from fastapi import Depends
from psycopg2.extras import RealDictCursor
from dotenv import load_dotenv
from contextlib import asynccontextmanager
from routers import avpRouter, unitRouter, itemRouter
import utils.globals as globals

load_dotenv()


@asynccontextmanager
async def lifespan(app: FastAPI):
    app.state.conn = globals.connPool.getConnection()
    yield
    globals.connPool.pool.putconn(app.state.conn)

def getDB():
    db = app.state.conn
    cursor = db.cursor(cursor_factory=RealDictCursor)
    app.state.cursor = cursor
    yield
    app.state.cursor.close()

app = FastAPI(lifespan=lifespan,
              dependencies=[Depends(getDB)]
              )

app.include_router(avpRouter.router)

app.include_router(unitRouter.router)

app.include_router(itemRouter.router)

@app.get("/")
async def pingApi():
    return {"ping": "pong"}