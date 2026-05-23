# Hackathon BQIA - Backend Go

Backend en Go con Clean Architecture, GORM y PostgreSQL.

## Requisitos

- Go 1.24.2+
- PostgreSQL

## Estructura

```
├── commands/                 # Scripts de ejecución
├── cmd/migrate/              # Migraciones GORM
├── internal/features/          # Módulos por funcionalidad
├── internal/infrastructure/  # Config, DB, HTTP
├── internal/shared/          # Utilidades compartidas
├── .env.dev                  # Configuración desarrollo
├── .env.prod                 # Configuración producción
└── main.go
```

## Inicio rápido

1. Copia y ajusta las variables en `.env.dev`
2. Crea la base de datos PostgreSQL:

```bash
createdb hackathon_bqia
```

3. Ejecuta en desarrollo:

```bash
chmod +x commands/*.sh
./commands/dev.sh
```

## Comandos

```bash
./commands/dev.sh      # Desarrollo
./commands/start.sh   # Producción
./commands/migrate.sh # Solo migraciones
```

## API

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/users` | Listar usuarios |
| POST | `/api/v1/users` | Crear usuario |
| GET | `/api/v1/users/{id}` | Obtener usuario |

## Stack

- Go 1.24.2 + GORM + PostgreSQL
- Gorilla Mux
