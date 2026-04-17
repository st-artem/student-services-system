from sqlalchemy import Column, Integer, String, Boolean
from sqlalchemy.orm import declarative_base

Base = declarative_base()

class Student(Base):
    __tablename__ = "students"

    id = Column(Integer, primary_key=True, index=True)
    first_name = Column(String, nullable=False)
    last_name = Column(String, nullable=False)
    email = Column(String, unique=True, index=True, nullable=False)
    group_name = Column(String, nullable=False)
    is_active = Column(Boolean, default=True)
    enrollment_year = Column(Integer, nullable=False) # Завдання 1: нове поле