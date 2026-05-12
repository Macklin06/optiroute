from fastapi import FastAPI
from app.routers import predict
from app.subscriber import start_subscriber

app = FastAPI(
    title="OptiRoute Predictor",
    description="ML demand prediction microservice",
    version="1.0.0",
)

app.include_router(predict.router, prefix="/api/v1", tags=["predictions"])


@app.on_event("startup")
async def startup_event():
    start_subscriber()


@app.get("/health")
async def health():
    return {"status": "ok", "service": "optiroute-predictor"}