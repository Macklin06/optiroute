from fastapi import APIRouter
from pydantic import BaseModel, Field

router = APIRouter()


class PredictionRequest(BaseModel):
    zone_id: str = Field(..., description="Zone identifier")
    hour_of_day: int = Field(..., ge=0, le=23)
    day_of_week: int = Field(..., ge=0, le=6)
    is_raining: bool = Field(default=False)
    current_orders: int = Field(..., ge=0)


class PredictionResponse(BaseModel):
    zone_id: str
    predicted_orders: float
    model_version: str


def calculate_prediction(req: PredictionRequest) -> float:
    base = 10.0

    if 11 <= req.hour_of_day <= 14:
        base *= 1.5
    elif 18 <= req.hour_of_day <= 21:
        base *= 1.8

    if req.day_of_week >= 5:
        base *= 1.3

    if req.is_raining:
        base *= 1.4

    base += req.current_orders * 0.5

    return round(base, 2)


@router.post("/predict_demand", response_model=PredictionResponse)
async def predict_demand(request: PredictionRequest):
    predicted = calculate_prediction(request)

    return PredictionResponse(
        zone_id=request.zone_id,
        predicted_orders=predicted,
        model_version="mock-v1",
    )