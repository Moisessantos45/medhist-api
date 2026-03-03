# 🐾 Veterinary Appointment API

REST API para la gestión de clínicas veterinarias. Permite administrar veterinarios, pacientes (animales), citas, vacunas e historial médico.

## 🛠️ Stack Tecnológico

| Tecnología       | Uso                                                 |
| ---------------- | --------------------------------------------------- |
| **Go**           | Lenguaje principal                                  |
| **Gin**          | Framework HTTP / router                             |
| **PostgreSQL**   | Base de datos relacional (GORM)                     |
| **Redis**        | Caché de sesiones y tokens                          |
| **Paseto**       | Autenticación (tokens)                              |
| **Email nativo** | Confirmación de cuenta y recuperación de contraseña |
| **godotenv**     | Variables de entorno                                |

---

## 🏛️ Arquitectura — Layered Architecture

El proyecto sigue una **Layered Architecture** donde cada feature está encapsulada en su propio paquete con tres capas bien definidas:

```
internal/
├── features/
│   ├── appointments/
│   │   ├── handler.go      ← Capa HTTP: recibe requests, valida y responde
│   │   ├── service.go      ← Capa de negocio: lógica, reglas, Redis
│   │   └── repository.go   ← Capa de datos: queries a PostgreSQL
│   ├── patients/
│   ├── veterinarians/
│   ├── vaccinations/
│   ├── medical_records/
│   └── auth/
├── pkg/
│   ├── models/             ← Entidades, interfaces Repository y UseCase
│   ├── middleware/         ← Auth middleware, rate limiter, cleanup
│   ├── generateJwt.go
│   ├── sendEmail.go
│   ├── brcrypt.go
│   └── validators.go
├── routes/                 ← Registro de rutas e inyección de dependencias
└── templates/              ← Plantillas HTML para emails
config/
├── db/                     ← Conexión y migración PostgreSQL
├── redis.go                ← Inicialización de Redis
└── email.go                ← Configuración SMTP
```

### Flujo de una petición

```
Request → Router → Middleware (Auth) → Handler → Service → Repository → PostgreSQL/Redis
```

Las dependencias se inyectan desde `routes/` siguiendo el principio de inversión de dependencias mediante interfaces definidas en `pkg/models/`.

---

## 📦 Modelos principales

### Veterinarian

```
ID, Name, Email, Password, Phone, Website, Token, EmailConfirmed
```

> Tiene relación 1:N con Patients, Appointments, MedicalRecords y Vaccinations.

### Patient (mascota)

```
ID, Name, Owner, OwnerEmail, OwnerPhone, Symptoms, Status (active/inactive)
VeterinarianID (FK)
```

### Appointment (cita)

```
ID, Date, Reason, Status (scheduled/completed/canceled), Notes
PatientID (FK), VeterinarianID (FK)
```

### Vaccination (vacuna)

```
ID, Type, Date, NextDueDate, Status (completed/pending/canceled)
PatientID (FK), VeterinarianID (FK)
```

### MedicalRecord (historial médico)

```
ID, VisitDate, Diagnosis, Treatment, Prescription, WeightKg, TemperatureC, Notes
PatientID (FK), VeterinarianID (FK)
```

---

## 🔌 Endpoints

Base URL: `/api/v1`

### 🔐 Auth — `/api/v1/auth`

| Método | Ruta               | Auth | Descripción                      |
| ------ | ------------------ | ---- | -------------------------------- |
| POST   | `/login`           | ❌   | Iniciar sesión                   |
| POST   | `/forgot-password` | ❌   | Enviar email de recuperación     |
| GET    | `/confirm-account` | ✅   | Confirmar cuenta por email       |
| GET    | `/session`         | ✅   | Obtener sesión activa            |
| POST   | `/logout`          | ✅   | Cerrar sesión                    |
| POST   | `/reset-password`  | ✅   | Restablecer contraseña con token |
| POST   | `/change-password` | ✅   | Cambiar contraseña               |

---

### 👨‍⚕️ Veterinarians — `/api/v1/veterinarian`

| Método | Ruta                     | Auth | Descripción                      |
| ------ | ------------------------ | ---- | -------------------------------- |
| POST   | `/veterinarian/register` | ❌   | Registrar veterinario            |
| GET    | `/veterinarian`          | ✅   | Listar veterinarios (paginado)   |
| GET    | `/veterinarian/:id`      | ✅   | Obtener veterinario por ID       |
| GET    | `/veterinarian/session`  | ✅   | Obtener veterinario de la sesión |
| PUT    | `/veterinarian/:id`      | ✅   | Actualizar veterinario           |
| DELETE | `/veterinarian/:id`      | ✅   | Eliminar veterinario             |

---

### 🐶 Patients — `/api/v1/patient`

| Método | Ruta                  | Auth | Descripción                      |
| ------ | --------------------- | ---- | -------------------------------- |
| GET    | `/patient`            | ✅   | Listar pacientes (paginado)      |
| GET    | `/patient/:id`        | ✅   | Obtener paciente por ID          |
| POST   | `/patient`            | ✅   | Crear paciente                   |
| PUT    | `/patient/:id`        | ✅   | Actualizar paciente              |
| PATCH  | `/patient/:id/status` | ✅   | Cambiar estado (active/inactive) |
| DELETE | `/patient/:id`        | ✅   | Eliminar paciente                |

---

### 📅 Appointments — `/api/v1/appointment`

| Método | Ruta                                                             | Auth | Descripción             |
| ------ | ---------------------------------------------------------------- | ---- | ----------------------- |
| GET    | `/appointment/patient/:patient_id/veterinarian/:veterinarian_id` | ✅   | Listar citas (paginado) |
| GET    | `/appointment/:id`                                               | ✅   | Obtener cita por ID     |
| POST   | `/appointment`                                                   | ✅   | Crear cita              |
| PUT    | `/appointment/:id`                                               | ✅   | Actualizar cita         |
| DELETE | `/appointment/:id`                                               | ✅   | Eliminar cita           |

---

### 💉 Vaccinations — `/api/v1/vaccination`

| Método | Ruta                      | Auth | Descripción               |
| ------ | ------------------------- | ---- | ------------------------- |
| GET    | `/vaccination`            | ✅   | Listar vacunas (paginado) |
| GET    | `/vaccination/:id`        | ✅   | Obtener vacuna por ID     |
| POST   | `/vaccination`            | ✅   | Registrar vacuna          |
| PUT    | `/vaccination/:id`        | ✅   | Actualizar vacuna         |
| PATCH  | `/vaccination/:id/status` | ✅   | Cambiar estado            |
| DELETE | `/vaccination/:id`        | ✅   | Eliminar vacuna           |

---

### 📋 Medical Records — `/api/v1/medical-record`

| Método | Ruta                                                                | Auth | Descripción                 |
| ------ | ------------------------------------------------------------------- | ---- | --------------------------- |
| GET    | `/medical-record/patient/:patient_id/veterinarian/:veterinarian_id` | ✅   | Listar historial (paginado) |
| GET    | `/medical-record/:id`                                               | ✅   | Obtener registro por ID     |
| POST   | `/medical-record`                                                   | ✅   | Crear registro médico       |
| PUT    | `/medical-record/:id`                                               | ✅   | Actualizar registro         |
| DELETE | `/medical-record/:id`                                               | ✅   | Eliminar registro           |

---

## ⚙️ Variables de entorno

Crea un archivo `.env` en la raíz del proyecto:

```env
# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=your_db_name

# Redis
REDIS_ADDR=localhost:PORT
REDIS_PASSWORD=your_redis_password

# JWT / Paseto
PASETO_SECRET_KEY=your_32_char_secret_key

# Email (SMTP)
EMAIL_HOST=smtp.example.com
EMAIL_PORT=587
EMAIL_USER=your@email.com
EMAIL_PASSWORD=your_email_password
```

---

## 🚀 Ejecutar el proyecto

```bash
# Instalar dependencias
go mod tidy

# Ejecutar
go run main.go
```

El servidor inicia en el puerto **`:4101`** con graceful shutdown habilitado.

---

## 🔒 Autenticación

Todas las rutas protegidas requieren un token **Paseto** en el header:

```
Authorization: Bearer <token>
```

Los tokens se almacenan en **Redis** para permitir invalidación inmediata al hacer logout.

---

## 📨 Servicio de Email

La API utiliza un servicio de email nativo (SMTP) para:

- **Confirmación de cuenta** al registrar un veterinario
- **Recuperación de contraseña** mediante token enviado al correo

Las plantillas HTML de los emails se encuentran en `internal/templates/`.

---

## 📄 Características adicionales

- ✅ Paginación en todos los listados
- ✅ Compresión GZIP en todas las respuestas
- ✅ CORS configurado
- ✅ Graceful shutdown (30s timeout)
- ✅ Validación de inputs en la capa de dominio (models)
- ✅ Interfaces para desacoplamiento (Repository + UseCase)
