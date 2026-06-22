from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.routers import predict
from app.subscriber import start_subscriber

app = FastAPI(
    title="OptiRoute Predictor",
    description="ML demand prediction microservice",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(predict.router, prefix="/api/v1", tags=["predictions"])

@app.on_event("startup")
async def startup_event():
    start_subscriber()

@app.get("/health")
async def health():
    return {"status": "ok", "service": "optiroute-predictor"}