from pydantic import BaseModel, EmailStr, Field

class StudentBase(BaseModel):
    first_name: str
    last_name: str
    email: EmailStr
    group_name: str
    is_active: bool = True
    # Завдання 1: валідація року вступу
    enrollment_year: int = Field(ge=2000, le=2030, description="Рік вступу має бути між 2000 та 2030") 

class StudentCreate(StudentBase):
    pass

class StudentUpdate(BaseModel):
    # Завдання 4: всі поля опціональні для часткового оновлення
    first_name: str | None = None
    last_name: str | None = None
    email: EmailStr | None = None
    group_name: str | None = None
    is_active: bool | None = None
    enrollment_year: int | None = Field(default=None, ge=2000, le=2030)

class StudentResponse(StudentBase):
    id: int

    class Config:
        from_attributes = True