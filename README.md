# 🚚 Picka Backend

**Picka** is a last-mile logistics-as-a-service platform providing delivery APIs for vendors and real-time tracking tools for riders and customers.

This repo contains all backend microservices used by Picka, built using **Go**, **PostgreSQL**, **NATS**, and **Docker**.

---

## 🧱 Microservices

| Service             | Description |
|---------------------|-------------|
| `auth-service`      | Handles signup, login, JWT, and user management |
| `vendor-service`    | Manages vendor profile, products, and delivery requests |
| `shipment-service`  | Handles shipment creation, status tracking, and vendor-rider coordination |
| `rider-service`     | Manages rider onboarding, status, and assignments |
| `notification-service` | Listens to NATS events and triggers future notifications |
| `api-gateway`       | NGINX proxy for routing all incoming requests to services |

---

## 🧪 Tech Stack

- Go (Golang)
- PostgreSQL
- NATS Messaging
- Docker + Docker Compose
- NGINX API Gateway

---

## 🚀 Getting Started

1. Clone the repo:
   ```bash
   git clone https://github.com/hephi-io/picka-backend.git
   cd picka-backend
