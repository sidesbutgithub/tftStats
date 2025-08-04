from fastapi import APIRouter, Query, Request
from typing import Annotated

router = APIRouter(
    prefix="/listItems",
    tags=["listItems"],
    responses={404: {"Msg":"Resource Not Found"}},
)

@router.get("/")
async def getAllItems(req: Request, items: Annotated[list[str], Query()]=[]):
    db = req.app.state.cursor
    db.execute('''
               WITH itemPlacements AS(
               SELECT placement, UNNEST(items) AS itemName
               FROM units
               WHERE (
               items @> %s::varchar[]))
               SELECT AVG(placement) AS AVP, count(itemname), itemname
               FROM itemPlacements
               GROUP BY itemname
               ORDER BY AVG(placement)
               ''', (items, ))
    return db.fetchall()

@router.get("/{unitname}")
async def getItemsForUnit(req: Request, unitname: str, starLevel: int = 0, items: Annotated[list[str], Query()] = []):
    db = req.app.state.cursor
    if starLevel == 0:
        db.execute('''
                   WITH itemPlacements AS(
                   SELECT placement, UNNEST(items) AS itemName
                   FROM units
                   WHERE (
                   unitname = %s
                   AND items @> %s::varchar[])
                   )
                   SELECT AVG(placement) AS AVP, count(itemname), itemname
                   FROM itemPlacements
                   GROUP BY itemname
                   ORDER BY AVG(placement)
                   ''', (unitname, items,))
        return db.fetchall()
    db.execute('''
               WITH itemPlacements AS(
               SELECT placement, UNNEST(items) AS itemName
               FROM units WHERE (
               unitname = %s AND 
               starlevel = %s AND
               items @> %s::varchar[])
               )
               SELECT AVG(placement) AS AVP, count(itemname), itemname
               FROM itemPlacements
               GROUP BY itemname
               ORDER BY AVG(placement)
               ''', (unitname, starLevel, items,))
    return db.fetchall()