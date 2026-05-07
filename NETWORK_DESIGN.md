# NETWORK_DESIGN.md

## Мережева топологія
Вся система розгорнута в ізольованій Docker-мережі `prois-net`. Прямий доступ до мікросервісів з хост-машини обмежений (використовується `expose`). Єдина точка входу — `prois-nginx-proxy` на порту 8080.

## Таблиця сервісів та портів
| Сервіс | Внутрішній порт | Зовнішній URL через Proxy |
|---|---|---|
| student-service | 8000 | `/api/v1/students/` |
| order-service | 8080 | `/api/v1/orders/` |
| wallet-service | 8080 | `/api/v1/wallets/` |
| marketplace-service | 8080 | `/api/v1/marketplace/` |

## Політики таймаутів
- `proxy_connect_timeout`: 5s
- `proxy_read_timeout`: 15s

## Трасування (Observability)
Кожен запит маркується заголовком `X-Correlation-Id`. Якщо клієнт не передав його, Nginx генерує ID автоматично.