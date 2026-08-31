# Toll Plazas Between Two Indian Pincodes - Backend Service (Go)

A high-performance Golang backend service that calculates driving highway routes between two Indian postal PIN codes, identifies all toll plazas located along the route, computes cumulative distances from the starting origin, and serves responses with sub-millisecond Redis caching.

---

## Architecture & System Overview

```
                           +-------------------------------------+
                           |            HTTP Client              |
                           |   (Postman / Web App / Mobile)      |
                           +------------------+------------------+
                                              |
                                              | POST /api/v1/toll-plazas
                                              v
                           +-------------------------------------+
                           |          Gin HTTP Handler           |
                           |  (Validation, CORS, Recovery, Log)  |
                           +------------------+------------------+
                                              |
                                              v
                           +-------------------------------------+
                           |            Redis Cache              | <----+ (Cache Hit)
                           |      (route:{source}:{dest})        |
                           +------------------+------------------+
                                              | (Cache Miss)
                                              v
+-------------------------------------------------------------------------------------------+
|                                    Toll Service                                           |
|                                                                                           |
|  1. Geocoding Engine                                                                      |
|     - Check PostgreSQL `pincode_cache` table                                              |
|     - Fallback: OSM Nominatim API (`/search?postalcode={pin}&country=India`)              |
|                                                                                           |
|  2. Routing Engine                                                                        |
|     - Calls OSRM Driving Engine (`/route/v1/driving/{srcLon},{srcLat};{dstLon},{dstLat}`) |
|     - Extracts route polyline coordinates & calculates cumulative distances               |
|                                                                                           |
|  3. Spatial Toll Projection & Deduplication                                               |
|     - PostGIS / Bounding Box spatial candidate query                                      |
|     - Cross-track perpendicular projection (<= 500m buffer)                               |
|     - Cumulative distance from source calculation                                         |
|     - Twin/opposite carriageway deduplication (<= 1km)                                    |
+-------------------------------------------------------------------------------------------+
                                              |
                                              v
                               +-----------------------------+
                               |     PostgreSQL + PostGIS    |
                               |  - `toll_plazas` (GiST)     |
                               |  - `pincode_cache`          |
                               +-----------------------------+
```

---

## Features

- **Accurate Spatial Algorithm**: Projects candidate toll plazas onto polyline highway segments using cross-track distance calculations with a customizable perpendicular buffer (default 500m) and deduplicates opposite-lane plazas.
- **Resilient Geocoding**: Resolves any 6-digit Indian PIN code to exact coordinates with database caching and OSM Nominatim fallback.
- **Fast Ingestion & Upsert**: Idempotently seeds 2,387+ toll plazas from `toll_plaza_india.csv` on server startup or via standalone CLI.
- **Redis Caching**: Caches route results with configurable TTL for instantaneous repeated queries.
- **Strict Error Handling & Spec Compliance**: Formats responses, empty lists, and error JSON exactly as mandated by the assignment specification.
- **Clean Layered Architecture**: Decoupled handlers, services, repositories, geometry utilities, and DTOs.
- **Docker Compose Ready**: One-command spin-up for Go Service, PostGIS PostgreSQL, and Redis.

---

## File Structure

```
toll_plaza/
├── cmd/
│   ├── server/
│   │   └── main.go                         # Server entrypoint
│   └── seeder/
│       └── main.go                         # Standalone CSV seeder CLI
├── internal/
│   ├── cache/
│   │   └── redis.go                        # Redis caching layer with fallback
│   ├── config/
│   │   └── config.go                       # Environment configuration loader
│   ├── domain/
│   │   ├── errors.go                       # Domain error definitions
│   │   ├── models.go                       # Domain models (TollPlaza, Point, Route)
│   │   ├── request.go                      # Request DTOs
│   │   └── response.go                     # Response DTOs
│   ├── handler/
│   │   ├── health_handler.go               # GET /health endpoint
│   │   ├── middleware.go                   # Logger, CORS, Recovery middlewares
│   │   └── toll_handler.go                 # POST /api/v1/toll-plazas endpoint
│   ├── repository/
│   │   └── postgres/
│   │       ├── db.go                       # DB connection & auto migrations
│   │       ├── pincode_repo.go             # Pincode location cache repository
│   │       └── toll_repo.go                # PostGIS spatial toll repository & upsert
│   ├── service/
│   │   ├── geocoding_service.go            # Pincode -> Coordinate geocoding
│   │   ├── routing_service.go              # OSRM Driving route engine
│   │   ├── seeder_service.go               # CSV parsing & batch upsert
│   │   ├── spatial_service.go              # Polyline cross-track projection & deduplication
│   │   └── toll_service.go                 # Main orchestrator service
│   └── validator/
│       └── validator.go                    # Indian PIN code validator
├── migrations/
│   └── 001_init.sql                        # PostGIS schema & GiST index
├── pkg/
│   └── geometry/
│       ├── haversine.go                    # Great-circle distance calculations
│       └── polyline.go                     # Point-to-segment projection & bounding box
├── postman/
│   └── toll_plaza_api.postman_collection.json # Ready-to-import Postman collection
├── .env.example                            # Sample configuration
├── docker-compose.yml                      # Full containerized stack
├── Dockerfile                              # Multi-stage Go Dockerfile
├── go.mod
├── go.sum
├── toll_plaza_india.csv                    # Dataset of 2,387 toll plazas
└── README.md
```

---

## Quick Start Guide

### Option 1: Run with Docker Compose (Recommended)

1. Ensure Docker Desktop is running.
2. Run:
   ```bash
   docker-compose up --build
   ```
3. The API will start at `http://localhost:8080`.

---

### Option 2: Run Locally with Go

1. **Start PostgreSQL with PostGIS & Redis**:
   ```bash
   docker run -d --name toll_postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=toll_plaza_db postgis/postgis:15-3.3
   docker run -d --name toll_redis -p 6379:6379 redis:7-alpine
   ```

2. **Configure Environment**:
   ```bash
   cp .env.example .env
   ```

3. **Run the Application**:
   ```bash
   go run cmd/server/main.go
   ```

4. *(Optional)* Run Seeder manually:
   ```bash
   go run cmd/seeder/main.go
   ```

---

## API Documentation

### 1. Calculate Toll Plazas Between Two Pincodes

- **Endpoint**: `POST /api/v1/toll-plazas`
- **Headers**: `Content-Type: application/json`

#### Request Body:
```json
{
  "sourcePincode": "560064",
  "destinationPincode": "411045"
}
```

#### Success Response (`200 OK`):
```json
{
  "route": {
    "sourcePincode": "560064",
    "destinationPincode": "411045",
    "distanceInKm": 855
  },
  "tollPlazas": [
    {
      "name": "Navalgund Toll Plaza",
      "latitude": 15.53982,
      "longitude": 75.3614,
      "distanceFromSource": 412
    },
    {
      "name": "Hattargi Toll Plaza",
      "latitude": 16.1824,
      "longitude": 74.4521,
      "distanceFromSource": 540
    },
    {
      "name": "Khedshivapur Toll Plaza",
      "latitude": 18.3412,
      "longitude": 73.8546,
      "distanceFromSource": 828
    }
  ]
}
```

---

### 2. Error Responses

#### A. Invalid Pincode (`400 Bad Request`)
*When a pincode is not 6 digits, starts with 0, contains letters, or cannot be geocoded:*
```json
{
  "error": "Invalid source or destination pincode"
}
```

#### B. Same Source & Destination (`400 Bad Request`)
```json
{
  "error": "Source and destination pincodes cannot be the same"
}
```

#### C. No Toll Plazas Found (`200 OK`)
```json
{
  "route": {
    "sourcePincode": "110001",
    "destinationPincode": "110002",
    "distanceInKm": 4
  },
  "tollPlazas": []
}
```

---

## Running Tests

Run all unit tests with code coverage:
```bash
go test -v ./... -cover
```

---

## Postman Testing

Import [`postman/toll_plaza_api.postman_collection.json`](file:///d:/Golang_Projects/toll_plaza/postman/toll_plaza_api.postman_collection.json) directly into Postman to test:
1. **Benchmark test** (Bengaluru `560064` -> Pune `411045`).
2. **Long distance test** (Delhi `110001` -> Bengaluru `560001`).
3. **Invalid pincode validation**.
4. **Same pincode validation**.
5. **Health check**.
