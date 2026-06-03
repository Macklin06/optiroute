from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field
import pickle
import numpy as np
import shap
from pathlib import Path

router = APIRouter()

BASE_DIR = Path(__file__).parent.parent.parent
MODELS_DIR = BASE_DIR / "models"

with open(MODELS_DIR / "xgboost_demand_model.pkl", "rb") as f:
    model = pickle.load(f)

with open(MODELS_DIR / "label_encoder.pkl", "rb") as f:
    le = pickle.load(f)

explainer = shap.TreeExplainer(model)

FEATURES = [
    "hour_of_day", "day_of_week", "is_weekend", "is_raining",
    "base_zone_demand", "prev_hour_orders", "zone_encoded",
    "is_lunch_hour", "is_dinner_hour", "is_breakfast_hour",
    "is_late_night", "hour_sin", "hour_cos", "rain_weekend"
]

ZONE_BASE_DEMAND = {
    "koramangala": 18, "indiranagar": 15, "whitefield": 12,
    "marathahalli": 14, "hsr_layout": 13, "jp_nagar": 11,
    "electronic_city": 10, "hebbal": 9,
}

print(" XGBoost model loaded successfully")

# Request and Response models
class PredictionRequest(BaseModel):
    zone_id: str = Field(..., description="Zone identifier")
    hour_of_day: int = Field(..., ge=0, le=23)
    day_of_week: int = Field(..., ge=0, le=6)
    is_raining: bool = Field(default=False)
    current_orders: int = Field(..., ge=0)

class SHAPExplanation(BaseModel):
    feature: str
    contribution: float

class PredictionResponse(BaseModel):
    zone_id: str
    predicted_orders: float
    model_version: str
    top_factors: list[SHAPExplanation]

# Feature Engineering
def build_features(req: PredictionRequest) -> np.ndarray:
    is_weekend = 1 if req.day_of_week >= 5 else 0
    is_raining = 1 if req.is_raining else 0
    base_demand = ZONE_BASE_DEMAND.get(req.zone_id, 12)

    try:
        zone_encoded = int(le.transform([req.zone_id])[0])
    except ValueError:
        raise HTTPException(
            status_code=422,
            detail=f"Unknown zone: {req.zone_id}. Valid zones: {list(ZONE_BASE_DEMAND.keys())}"
        )

    is_lunch      = 1 if 11 <= req.hour_of_day <= 13 else 0
    is_dinner     = 1 if 19 <= req.hour_of_day <= 21 else 0
    is_breakfast  = 1 if  7 <= req.hour_of_day <=  9 else 0
    is_late_night = 1 if req.hour_of_day <= 5 else 0

    hour_sin = np.sin(2 * np.pi * req.hour_of_day / 24)
    hour_cos = np.cos(2 * np.pi * req.hour_of_day / 24)

    rain_weekend = is_raining * is_weekend

    return np.array([[
        req.hour_of_day, req.day_of_week, is_weekend, is_raining,
        base_demand, req.current_orders, zone_encoded,
        is_lunch, is_dinner, is_breakfast,
        is_late_night, hour_sin, hour_cos, rain_weekend
    ]])

# Endpoint
@router.post("/predict_demand", response_model = PredictionResponse)
async def predict_demand(request: PredictionRequest):
    features = build_features(request)

    predicted = float(model.predict(features)[0])
    predicted = max(0.0, round(predicted,2))

    shap_vals = explainer.shap_values(features)[0]
    shap_pairs = sorted(
        zip(FEATURES, shap_vals),
        key=lambda x: abs(x[1]),
        reverse=True
    )
    top_factors = [
        SHAPExplanation(feature=f,contribution=round(float(v),3))
        for f,v in shap_pairs[:3]
    ]

    return PredictionResponse(
        zone_id=request.zone_id,
        predicted_orders=predicted,
        model_version="xgboost-v1",
        top_factors=top_factors,
    )