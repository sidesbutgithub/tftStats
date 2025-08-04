from fastapi import APIRouter, Query, Request
from typing import Annotated

router = APIRouter(
    prefix="/builds",
    tags=["builds"],
    responses={404: {"Msg":"Resource Not Found"}},
)

@router.get("/")
async def getAllBuilds(req: Request, unitName: str = "", starLevel: int = 0, items: Annotated[list[str], Query()]=[], minGames: int = 10):
    db = req.app.state.cursor
    if unitName != "":
        if starLevel != 0:
            db.execute('''
                       SELECT * FROM(
                       SELECT AVG(placement) AS avp, count(id) as games, unitname, starlevel, items
                       FROM units
                       WHERE (unitname = %s
                       AND starlevel = %s
                       AND items @> %s::varchar[])
                       GROUP BY unitname, starlevel, items
                       ORDER BY avp)
                       WHERE games >= %s
                       ''', (unitName, starLevel, items, minGames,))
            return db.fetchall()
        else:
            db.execute('''
                       SELECT * FROM(
                       SELECT AVG(placement) AS avp, count(id) as games, unitname, starlevel, items
                       FROM units
                       WHERE (unitname = %s
                       AND items @> %s::varchar[])
                       GROUP BY unitname, starlevel, items
                       ORDER BY avp)
                       WHERE games >= %s
                       ''', (unitName, items, minGames,))
            return db.fetchall()

    db.execute('''
               SELECT * FROM(
               SELECT AVG(placement) AS avp, count(id) as games, unitname, starlevel, items
               FROM units
               WHERE items @> %s::varchar[]
               GROUP BY unitname, starlevel, items
               ORDER BY avp)
               WHERE games >= %s
               ''', (items, minGames, ))
    return db.fetchall()
