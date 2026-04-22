import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI, Depends, Request
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import text
from app.api.routes import router
from app.db.database import init_db, get_session
import time
from fastapi import Request
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from starlette.exceptions import HTTPException as StarletteHTTPException


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

@app.exception_handler(StarletteHTTPException)
async def custom_http_exception_handler(request: Request, exc: StarletteHTTPException):
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "error": {
                "code": exc.status_code,
                "message": exc.detail,
                "path": request.url.path,
                "timestamp": int(time.time())
            }
        }
    )

@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    return JSONResponse(
        status_code=422,
        content={
            "error": {
                "code": 422, 
                "message": "Validation Error", 
                "details": exc.errors(),
                "path": request.url.path,
                "timestamp": int(time.time())
            }
        }
    )

app.include_router(router)

@app.get("/health", tags=["System"])
async def health_check(db: AsyncSession = Depends(get_session)):
    db_status = "ok"
    try:
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