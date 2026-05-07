# Етап 1: Збірка (Build)
FROM maven:3.9.6-eclipse-temurin-21-alpine AS builder
WORKDIR /app
COPY pom.xml .
# Завантажуємо залежності (кешування)
RUN mvn dependency:go-offline
COPY src ./src
# Збираємо jar файл
RUN mvn clean package -DskipTests

# Етап 2: Запуск (Run)
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=builder /app/target/marketplace-service-0.0.1-SNAPSHOT.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]