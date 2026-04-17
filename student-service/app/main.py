import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI, Depends, HTTPException, Request
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import text
from app.api.routes import router
from app.db.database import init_db, get_session

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Starting up Student Service and initializing DB...")
    await init_db()
    yield
    logger.info("Shutting down Student Service...")

app = FastAPI(
    title="Student Service API",
    description="Microservice for managing students",
    version="1.0.0",
    lifespan=lifespan
)

# Завдання 8: Єдина структура обробки помилок
@app.exception_handler(HTTPException)
async def custom_http_exception_handler(request: Request, exc: HTTPException):
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": {"code": exc.status_code, "message": exc.detail}}
    )

@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    return JSONResponse(
        status_code=422,
        content={
            "error": {
                "code": 422, 
                "message": "Validation Error", 
                "details": exc.errors()
            }
        }
    )

app.include_router(router)

# Завдання 7: Розширений Health Endpoint з перевіркою БД
@app.get("/health", tags=["System"])
async def health_check(db: AsyncSession = Depends(get_session)):
    db_status = "ok"
    try:
        # Перевірка пінгу БД
        await db.execute(text("SELECT 1"))
    except Exception as e:
        logger.error(f"Database health check failed: {e}")
        db_status = "error"
        
    status_code = 200 if db_status == "ok" else 503
    
    return JSONResponse(
        status_code=status_code,
        content={
            "status": "ok" if db_status == "ok" else "degraded", 
            "service": "student-service",
            "database": db_status
        }
    )