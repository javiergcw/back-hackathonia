# Yakka Backend - Arquitectura

## 📁 Estructura del Proyecto
```
yakka-go-exagone-back/
├── commands/                 # Scripts de ejecución
│   ├── dev.sh               # Desarrollo
│   ├── start.sh             # Producción
│   └── migrate.sh           # Migraciones
├── internal/features/        # Módulos de funcionalidades
│   ├── auth/                # Autenticación
│   │   ├── user/            # Usuarios
│   │   ├── builder_profiles/ # Perfiles constructor
│   │   └── labour_profiles/  # Perfiles trabajador
│   └── users/               # Usuarios (legacy)
├── internal/infrastructure/  # Infraestructura
│   ├── config/              # Configuración
│   ├── database/            # Base de datos
│   └── http/                # HTTP server
├── internal/shared/         # Código compartido
│   ├── errors/              # Manejo de errores
│   ├── response/            # Respuestas HTTP
│   └── validation/          # Validaciones
├── .env.dev                 # Configuración desarrollo
├── .env.prod                # Configuración producción
└── main.go                  # Punto de entrada
```

## 🏗️ Arquitectura
- **Clean Architecture** con separación de capas
- **Feature-based** por módulos
- **GORM** para migraciones automáticas
- **Scripts bash** para ejecución

## 🚀 Comandos
```bash
# Desarrollo
./commands/dev.sh

# Producción  
./commands/start.sh

# Migraciones
./commands/migrate.sh
```

## 🛠️ Stack
- **Go 1.24.2** + **GORM** + **PostgreSQL**
- **Gorilla Mux** (router)
- **Scripts bash** (sin Docker)
