from fastapi import APIRouter, Query, Request
from typing import Annotated

router = APIRouter(
    prefix="/listUnits",
    tags=["listUnits"],
    responses={404: {"Msg":"Resource Not Found"}},
)

@router.get("/")
async def getAllUnits(req: Request, items: Annotated[list[str], Query()] = []):
    db = req.app.state.cursor
    db.execute('''
               SELECT AVG(placement) AS avp, unitname, starlevel, COUNT(units)
               FROM units
               WHERE items @> %s::varchar[]
               GROUP BY unitname, starlevel
               ORDER BY avp
               ''', (items,))
    return db.fetchall()
