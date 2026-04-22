FROM golang:alpine
WORKDIR /app

# Копіюємо всю твою структуру папок у контейнер
COPY . .

# Docker сам ініціалізує проєкт і качає бібліотеки
RUN go mod init order-service
RUN go get -u github.com/gin-gonic/gin gorm.io/gorm gorm.io/driver/postgres github.com/redis/go-redis/v9 github.com/joho/godotenv
RUN go mod tidy

# Компілюємо код із правильної папки
RUN go build -o main ./cmd/api

CMD ["./main"]