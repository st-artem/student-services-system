from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy import desc, asc, func
from sqlalchemy.exc import IntegrityError
from app.db.database import get_session
from app.db.models import Student
from app.schemas.student import StudentCreate, StudentResponse, StudentUpdate


router = APIRouter(prefix="/students", tags=["Students"])

@router.get("/stats/summary")
async def get_stats(db: AsyncSession = Depends(get_session)):
    total_students = await db.scalar(select(func.count(Student.id)))
    
    group_stats_query = select(Student.group_name, func.count(Student.id)).group_by(Student.group_name)
    group_stats_result = await db.execute(group_stats_query)
    
    group_stats = [{"group": row[0], "count": row[1]} for row in group_stats_result.all()]
    
    return {
        "total_students": total_students,
        "by_group": group_stats
    }

@router.post("/", response_model=StudentResponse, status_code=201)
async def create_student(student: StudentCreate, db: AsyncSession = Depends(get_session)):
    new_student = Student(**student.model_dump())
    db.add(new_student)
    try:
        await db.commit()
        await db.refresh(new_student)
        return new_student
    except IntegrityError:
        await db.rollback()
        raise HTTPException(status_code=409, detail="Student with this email already exists")

@router.get("/", response_model=list[StudentResponse])
async def get_students(
    group_name: str | None = None,                 
    last_name_contains: str | None = None,       
    sort_by: str = "id",                           
    sort_order: str = "asc",                       
    skip: int = 0,                                 
    limit: int = 10,                               
    db: AsyncSession = Depends(get_session)
):
    query = select(Student)
    
    if group_name:
        query = query.where(Student.group_name == group_name)
    if last_name_contains:
        query = query.where(Student.last_name.ilike(f"%{last_name_contains}%"))

    sort_attr = getattr(Student, sort_by, Student.id)
    if sort_order == "desc":
        query = query.order_by(desc(sort_attr))
    else:
        query = query.order_by(asc(sort_attr))

    query = query.offset(skip).limit(limit)
    
    result = await db.execute(query)
    return result.scalars().all()

@router.get("/{student_id}", response_model=StudentResponse)
async def get_student(student_id: int, db: AsyncSession = Depends(get_session)):
    student = await db.get(Student, student_id)
    if not student:
        raise HTTPException(status_code=404, detail="Student not found")
    return student

@router.patch("/{student_id}", response_model=StudentResponse)
async def update_student(student_id: int, student_update: StudentUpdate, db: AsyncSession = Depends(get_session)):
    student = await db.get(Student, student_id)
    if not student:
        raise HTTPException(status_code=404, detail="Student not found")
    
    update_data = student_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(student, key, value)
        
    try:
        await db.commit()
        await db.refresh(student)
        return student
    except IntegrityError:
        await db.rollback()
        raise HTTPException(status_code=409, detail="Email is already in use by another student")

@router.delete("/{student_id}", status_code=204)
async def delete_student(student_id: int, db: AsyncSession = Depends(get_session)):
    student = await db.get(Student, student_id)
    if not student:
        raise HTTPException(status_code=404, detail="Student not found")
    await db.delete(student)
    await db.commit()