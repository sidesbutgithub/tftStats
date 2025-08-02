from fastapi import APIRouter, Query, Request
from typing import Annotated

router = APIRouter(
    prefix="/avp",
    tags=["avp"],
    responses={404: {"Msg":"Resource Not Found"}},
)

@router.get("/")
async def getAvp(request: Request, unitName: str = "", starLevel: int = 0, items: Annotated[list[str], Query()] = []):
    db = request.app.state.cursor

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

