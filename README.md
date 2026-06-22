# OptiRoute — Real-time Fleet Demand Prediction

A multi-service system that predicts short-term demand per city zone and visualizes live driver locations for fleet optimization.

**Tech Stack**: Go · Python · Vue 3 · XGBoost · Redis · PostgreSQL · Leaflet · FastAPI · Gin

---

## Quick Demo (Local)

### Prerequisites
- Docker & Docker Compose
- Git

### 1. Clone & Navigate
```bash
git clone <repo-url> optiroute
cd optiroute
```

### 2. Start Services
```bash
docker-compose up
```

This spins up:
- **Redis** (port 6379) — cache + pub/sub
- **PostgreSQL** (port 5432) — persistent store
- **Router** (port 8080) — Go API
- **Predictor** (port 8000) — Python ML service
- **Dashboard** (port 5173) — Vue frontend

### 3. Seed Demo Fleet & View Dashboard
In a new terminal:
```bash
# Start the demo fleet simulator (continuously updates 20 drivers every 2 seconds)
cd optiroute
bash demo_simulator.sh
```

Then open **http://localhost:5173** in your browser.

Click the **"Spawn demo fleet"** button in the sidebar to populate the map with 20 active drivers.

---

## Project Overview

**OptiRoute** predicts demand for last-mile delivery in real-time zones and helps operators pre-position drivers.

### Key Components

| Component | Tech | Port | Purpose |
|-----------|------|------|---------|
| **Router** | Go + Gin | 8080 | API for driver updates, active drivers, orders; Routes to Redis/Postgres |
| **Predictor** | FastAPI + XGBoost | 8000 | ML demand forecasting; Subscribes to driver updates via Redis pub/sub |
| **Dashboard** | Vue 3 + Leaflet | 5173 | Real-time map visualization; Shows predictions & driver locations |
| **Redis** | Cache + Pub/Sub | 6379 | Live driver state (TTL=60s), event notifications |
| **PostgreSQL** | Persistent DB | 5432 | History of driver updates & orders |

### Data Flow

```
Driver Update:
  Driver Simulator
    ↓ PUT /api/v1/drivers/location
  Router (Go)
    ├→ Redis SET (TTL 60s)
    ├→ Postgres INSERT
    └→ Redis PUBLISH (driver:updates)
      ↓
  Predictor (subscribes)
      ↓
  Dashboard (polls /drivers/active)
    → renders live markers

Demand Prediction:
  Dashboard
    ↓ POST /api/v1/predict_demand
  Predictor (XGBoost + SHAP)
    ↓
  Dashboard
    → displays zone demand bubbles
```

---

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed design documentation.

### System Diagram
For an interactive diagram, download and open [docs/system-diagram.excalidraw](docs/system-diagram.excalidraw) in **[Excalidraw](https://excalidraw.com)**.

---

## API Endpoints

### Router (Go) — http://localhost:8080

#### Driver Location
```bash
# Update a driver's location
curl -X PUT http://localhost:8080/api/v1/drivers/location \
  -H "Content-Type: application/json" \
  -d '{
    "driver_id": "driver_001",
    "latitude": 12.9352,
    "longitude": 77.6245
  }'

# Get all active drivers (from Redis cache)
curl http://localhost:8080/api/v1/drivers/active
```

#### Orders
```bash
# Create an order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "cust_123",
    "latitude": 12.9500,
    "longitude": 77.6300
  }'

# Get pending orders in a zone
curl http://localhost:8080/api/v1/orders/zone/zone_northeast
```

#### Health & Metrics
```bash
curl http://localhost:8080/health
curl http://localhost:8080/metrics  # Prometheus format
```

### Predictor (Python) — http://localhost:8000

```bash
# Predict demand for a zone
curl -X POST http://localhost:8000/api/v1/predict_demand \
  -H "Content-Type: application/json" \
  -d '{
    "zone_id": "koramangala",
    "hour_of_day": 14,
    "day_of_week": 3,
    "is_raining": false,
    "current_orders": 10
  }'

# Response example:
# {
#   "zone_id": "koramangala",
#   "predicted_orders": 18.45,
#   "model_version": "xgboost-v1",
#   "top_factors": [
#     {"feature": "hour_of_day", "contribution": 2.345},
#     {"feature": "is_lunch_hour", "contribution": 1.892},
#     {"feature": "base_zone_demand", "contribution": 0.756}
#   ]
# }
```

---

## Running Locally (Without Docker)

### Prerequisites per Service
- **Router**: Go 1.20+
- **Predictor**: Python 3.9+
- **Dashboard**: Node 20+
- **Redis**: Redis 6+
- **Postgres**: PostgreSQL 13+

### Router (Go)
```bash
cd router
go mod download
go run cmd/server/main.go
```

### Predictor (Python)
```bash
cd predictor
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000
```

### Dashboard (Vue)
```bash
cd dashboard
npm install
npm run dev
```

### Start Redis & Postgres (if not Docker)
```bash
# Redis
redis-server

# Postgres (assumes installed locally)
# Create database: createdb -U postgres optiroute
```

---

## Demo Simulator

Continuously updates 20 demo drivers with slight random movement:

```bash
cd optiroute
bash demo_simulator.sh
```

Or use the **"Spawn demo fleet"** button in the dashboard UI (it's built-in).

---

## Project Structure

```
optiroute/
├─ router/                     # Go backend (Gin, GORM, metrics)
│  ├─ cmd/server/main.go
│  ├─ internal/
│  │  ├─ handlers/
│  │  ├─ services/
│  │  ├─ models/
│  │  └─ config/
│  ├─ Dockerfile
│  └─ go.mod
│
├─ predictor/                  # Python backend (FastAPI, XGBoost, SHAP)
│  ├─ app/
│  │  ├─ main.py
│  │  ├─ subscriber.py
│  │  └─ routers/predict.py
│  ├─ models/
│  │  ├─ xgboost_demand_model.pkl
│  │  └─ label_encoder.pkl
│  ├─ requirements.txt
│  └─ Dockerfile
│
├─ dashboard/                  # Vue 3 frontend (Leaflet, Vite)
│  ├─ src/App.vue
│  ├─ package.json
│  └─ Dockerfile
│
├─ docs/
│  ├─ ARCHITECTURE.md
│  ├─ system-diagram.excalidraw
│  └─ system-diagram.png
│
├─ docker-compose.yml
├─ demo_simulator.sh
├─ .env.example
└─ README.md
```

---

## Deployment

### AWS (Free-Tier Friendly Setup)

#### Option 1: ECS Fargate + RDS + ElastiCache (Easiest)
1. Build and push images to **ECR**.
2. Create **ECS Service** for router and predictor.
3. Set up **RDS PostgreSQL** (t3.micro).
4. Set up **ElastiCache Redis** (cache.t3.micro).
5. Use **ALB** to route traffic to router service.
6. Deploy dashboard to **S3 + CloudFront**.

#### Option 2: Single EC2 Instance (Cheapest)
1. Launch **t3.micro EC2** (free tier, Ubuntu 22.04).
2. Install Docker & Docker Compose.
3. `docker-compose up`
4. Use **Elastic IP** + **Route53** for DNS.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for full deployment details.

---

## Development & Debugging

### View Logs
```bash
docker-compose logs -f router
docker-compose logs -f predictor
docker-compose logs -f dashboard
```

### Redis Inspection
```bash
redis-cli
KEYS driver:location:*
GET driver:location:driver_001
SUBSCRIBE driver:updates
```

### Postgres Inspection
```bash
psql -h localhost -U postgres -d optiroute
\dt  # List tables
SELECT * FROM driver_locations LIMIT 10;
```

---

## Interview Talking Points

1. **Multi-service Architecture**: Separation of concerns (router, predictor, frontend); scales independently.
2. **Redis + Postgres Trade-off**: Redis for speed + TTL + pub/sub; Postgres for durability + history.
3. **Circuit Breaker Pattern**: Graceful degradation — if Redis fails, system still works via Postgres fallback.
4. **ML Explainability**: SHAP shows which features influenced predictions (interpretability).
5. **Resilience**: TTL-based cleanup, conditional DB updates, failure mode handling.
6. **Scaling**: Cache cluster, read replicas, stateless predictor (horizontal scale).

---

## FAQs

**Q: Can I deploy this on AWS free tier?**  
A: Yes. Use t3.micro EC2, RDS free tier (t3.micro), and ElastiCache cache.t3.micro. Dashboard on S3 + CloudFront.

**Q: How do I add real drivers instead of simulators?**  
A: Integrate a driver mobile app that hits `PUT /api/v1/drivers/location` with periodic heartbeats.

**Q: What's the TTL for drivers?**  
A: 60 seconds. Drivers must heartbeat at least every 60s to remain "active". Tune based on your use case.

**Q: Is the dashboard secure?**  
A: Currently no auth. Add JWT + role-based access (admin, operator, driver) for production.

---

## Resume Summary (2–3 lines)

**OptiRoute** — Built a real-time fleet optimization system: Go router handles live driver ingestion with Redis cache + Postgres durability; Python FastAPI service predicts 30m demand via XGBoost + SHAP; Vue dashboard visualizes demand heatmap + driver locations. Implemented circuit breaker, pub/sub, and TTL-based cleanup for resilience.

---

## License
MIT

---

Ready to demo? Start with `docker-compose up` and open **http://localhost:5173** 🚀
