# Guía de Uso - ETL Construction with MongoDB

Guía rápida para ejecutar el ETL de MovieLens con MongoDB.

---

## 📑 Tabla de Contenidos

- [Requisitos](#-requisitos)
- [Instalación Rápida](#-instalación-rápida)
- [Ejecución del ETL](#-ejecución-del-etl)
- [Importación a MongoDB](#-importación-a-mongodb)
- [Verificación](#-verificación)

---

## 🔧 Requisitos

- **Go 1.21+**: https://go.dev/dl/
- **MongoDB 4.4+**: https://www.mongodb.com/try/download/community
- **TMDB API Key** (opcional): https://www.themoviedb.org/settings/api

---

## ⚡ Instalación Rápida

```powershell
# 1. Clonar repositorio
git clone https://github.com/PrograCyD/PC4_ETLConstructionWithMongoDB.git
cd PC4_ETLConstructionWithMongoDB

# 2. Instalar dependencias
go mod tidy

# 3. Configurar TMDB API (opcional)
Copy-Item .env.example .env
notepad .env  # Agregar: TMDB_API_KEY=tu_api_key_aqui
```

---

## 🚀 Ejecución del ETL

### Configuración por Defecto

**Sin especificar flags, el ETL usa:**
- ✅ CSV del directorio `data/` (usa `movies.csv`, NO `movies_test.csv`)
- ✅ Passwords **CON hash** bcrypt (producción segura)
- ❌ **SIN** datos externos de TMDB (solo datos locales)

### Casos de Uso Comunes

#### 1️⃣ Ejecución Completa (Todos los archivos)

```powershell
# Fase 1: Solo datos locales, CON hash (~1m 10s) - DEFAULT
go run .

# Fase 1: Solo datos locales, SIN hash (~1m) - MÁS RÁPIDO
go run . --hash-passwords=false

# Fase 2: Con TMDB API, SIN hash (~4-5 horas por rate limit)
go run . --fetch-external --hash-passwords=false

# Fase 2: Con TMDB API, CON hash (~4-5 horas + 10min) - MÁS COMPLETO
go run . --fetch-external
```

#### 2️⃣ Ejecución Selectiva (Solo archivos específicos)

```powershell
# Solo regenerar USERS (útil para cambiar hash on/off) - 8s
go run . --process-movies=false --process-ratings=false --process-similarities=false

# Solo MOVIES y RATINGS (~1m)
go run . --process-users=false --process-similarities=false

# Solo MOVIES con TMDB (~4-5 horas)
go run . --process-ratings=false --process-users=false --process-similarities=false --fetch-external

# Solo SIMILARITIES (~10s)
go run . --process-movies=false --process-ratings=false --process-users=false
```

#### 3️⃣ Pruebas Rápidas (Dataset pequeño)

```powershell
# 10 películas, SIN hash (~5s)
go run . --movies-file movies_test.csv --hash-passwords=false

# 10 películas, CON TMDB (~10s)
go run . --movies-file movies_test.csv --fetch-external --hash-passwords=false
```

### Flags Disponibles

#### Procesadores (default: todos activados)
```powershell
--process-movies=true/false         # Genera movies.ndjson
--process-ratings=true/false        # Genera ratings.ndjson  
--process-users=true/false          # Genera users.ndjson
--process-similarities=true/false   # Genera similarities.ndjson
```

#### Archivos CSV (default: usa archivos completos)
```powershell
--data-dir data                     # Directorio de CSVs
--movies-file movies.csv            # Archivo de películas
--ratings-file ratings.csv          # Archivo de ratings
--out-dir out                       # Directorio de salida
```

#### Configuración de Datos
```powershell
--hash-passwords=true/false         # Hash bcrypt (default: true)
--min-relevance 0.5                 # Relevancia genome tags (default: 0.5)
--top-genome-tags 10                # Max genome tags (default: 10)
--update-mappings=true/false        # Actualizar CSVs de mapeo (default: false)
```

#### TMDB API (default: desactivado)
```powershell
--fetch-external=true/false         # Obtener datos TMDB (default: false)
--tmdb-api-key "xxx"                # API key (lee de .env si no se especifica)
--tmdb-rate-limit 4                 # Req/s a TMDB (default: 4)
```

### Ejemplos Prácticos

```powershell
# Desarrollo: Sin hash, solo users, RÁPIDO
go run . --process-movies=false --process-ratings=false --process-similarities=false --hash-passwords=false

# Producción: Con hash, todo
go run . --hash-passwords=true

# Re-generar users con hash diferente
go run . --process-movies=false --process-ratings=false --process-similarities=false --hash-passwords=true

# Actualizar solo movies con TMDB (mantener ratings/users/similarities existentes)
go run . --process-ratings=false --process-users=false --process-similarities=false --fetch-external
```

---

## 📥 Importación a MongoDB

### 1. Iniciar MongoDB

```powershell
mongod --dbpath C:\data\db
```

### 2. Importar Colecciones

```powershell
# Variables
$DB = "movielens"
$OUT_DIR = "out"

# Importar todas las colecciones
mongoimport --db $DB --collection movies --file "$OUT_DIR\movies.ndjson"
mongoimport --db $DB --collection ratings --file "$OUT_DIR\ratings.ndjson"
mongoimport --db $DB --collection users --file "$OUT_DIR\users.ndjson"
mongoimport --db $DB --collection similarities --file "$OUT_DIR\similarities.ndjson"
```

### 3. Crear Índices (Recomendado)

```javascript
// En mongosh
use movielens

// Movies
db.movies.createIndex({ movieId: 1 })
db.movies.createIndex({ iIdx: 1 })
db.movies.createIndex({ title: "text" })

// Ratings
db.ratings.createIndex({ userId: 1, movieId: 1 })

// Users
db.users.createIndex({ userId: 1 }, { unique: true })
db.users.createIndex({ email: 1 }, { unique: true })

// Similarities
db.similarities.createIndex({ iIdx: 1 })
```

---

## ✅ Verificación

```javascript
// En mongosh
use movielens

// Contar documentos
db.movies.countDocuments()       // 62423
db.ratings.countDocuments()      // 25000095
db.users.countDocuments()        // 162541
db.similarities.countDocuments() // 30202

// Ver ejemplos
db.movies.findOne({ movieId: 1 })
db.users.findOne({ userId: 1 })
db.similarities.findOne({ movieId: 1 })

// Top películas
db.movies.find(
  { "ratingStats.count": { $gte: 1000 } }
).sort({ "ratingStats.average": -1 }).limit(10)
```

---

## 📋 Resumen de Respuestas

**P: ¿Si no especifico el CSV, usa el dataset completo o el de test?**  
R: ✅ Usa el **completo** (`movies.csv`), NO el de test.

**P: ¿Si no especifico flags, movies incluye datos externos?**  
R: ❌ NO. Por defecto `--fetch-external=false` (solo datos locales).

**P: ¿Si no especifico flags, los users se hashean?**  
R: ✅ SÍ. Por defecto `--hash-passwords=true` (seguro para producción).

---

## 📞 Soporte

- **README.md**: Documentación técnica completa
- **Repositorio**: https://github.com/PrograCyD/PC4_ETLConstructionWithMongoDB
- **Issues**: Reportar problemas en GitHub

**¡El ETL está listo para alimentar tu sistema de recomendaciones!** 🎬🍿