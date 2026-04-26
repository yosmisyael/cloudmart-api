A high-performance e-commerce backend engine built with Go, specifically architected to handle the demanding requirements of modern retail platforms. This project serves as a technical showcase for implementing Pragmatic Layered Architecture and scalable transaction management.

# Getting Started
## Prerequisites
- Go 1.21+
- PostgreSQL
- swag CLI (for docs generation)

## Installation
1. Clone this repo:
```sh
git clone https://github.com/yosmisyael/cloudmart-web-service.git
 ```
2. Setup environment variables:
```sh
cp .env.example .env
```
3. Generate swagger docs:
```sh
swag init -g cmd/api/main.go --parseDependency --parseInternal
```
4. Run the server:
```sh
go run cmd/api/main.go
```
