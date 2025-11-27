# ETL Construction with MongoDB - MovieLens Dataset

Sistema ETL (Extract, Transform, Load) desarrollado en Go para procesar el dataset MovieLens y generar colecciones enriquecidas en formato NDJSON para MongoDB, integrando datos externos de TMDB API.

---

## 📋 Tabla de Contenidos

- [Resumen del Dataset](#-resumen-del-dataset)
- [Diccionario de Datos](#-diccionario-de-datos)
- [Relevancia del Proyecto](#-relevancia-del-proyecto)
- [Arquitectura del Sistema](#-arquitectura-del-sistema)
- [Fundamentos Teóricos](#-fundamentos-teóricos)
- [Colecciones Generadas](#-colecciones-generadas)
- [Tecnologías Utilizadas](#-tecnologías-utilizadas)

---

## 📊 Resumen del Dataset

El dataset **MovieLens 25M** contiene **25,000,095 calificaciones** y **1,093,360 etiquetas** sobre **62,423 películas** evaluadas por **162,541 usuarios (1995–2019)**.

**Características principales:**
- Formato: **CSV con cabecera** (UTF-8, separador `,`)
- Archivos core: **ratings.csv**, **movies.csv**, **links.csv**, **tags.csv**
- Archivos complementarios: **genome-tags.csv**, **genome-scores.csv**
- Archivos de mapeo: **item_map.csv**, **user_map.csv** (generados en preprocesamiento PC3)
- Archivo de similitudes: **item_topk_cosine_conc.csv** (similitudes coseno k=20)

### Propósito de Archivos

| Archivo | ¿Para qué sirve en general? | ¿Para qué lo usamos en PC4? |
|---------|----------------------------|----------------------------|
| **ratings.csv** | Calificaciones usuario-película (base del filtrado colaborativo) | Generar colección `ratings` y calcular estadísticas agregadas para `movies` |
| **movies.csv** | Metadatos de películas (título, géneros) | Base de colección `movies` con enriquecimiento de tags y datos externos |
| **links.csv** | IDs externos (IMDB, TMDB) | Vincular con TMDB API para obtener posters, cast, sinopsis |
| **tags.csv** | Etiquetas libres asignadas por usuarios | Normalizar y rankear top 10 `userTags` por película |
| **genome-tags.csv** | 1,128 tags curados del Tag Genome | Diccionario para interpretar genome-scores |
| **genome-scores.csv** | Relevancia (0-1) de cada genome tag por película | Seleccionar top 10 `genomeTags` con relevancia ≥ 0.5 |
| **item_map.csv** | Mapeo movieId → iIdx (índice continuo 0..N-1) | Vincular películas con el modelo de recomendación (similitudes) |
| **user_map.csv** | Mapeo userId → uIdx (índice continuo 0..M-1) | Generar colección `users` con índices para el modelo |
| **item_topk_cosine_conc.csv** | Similitudes coseno k=20 pre-calculadas (PC3) | Generar colección `similarities` para recomendaciones |

---

## 📖 Diccionario de Datos

### ratings.csv

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `userId` | int | Identificador anónimo de usuario (1 a 162,541) |
| `movieId` | int | Identificador único de película (1 a 193,609, con gaps) |
| `rating` | float | Calificación de 0.5 a 5.0 (incrementos de 0.5) |
| `timestamp` | int64 | Momento de calificación en UNIX timestamp (UTC) |

**Notas**: 
- Ordenado por `userId`, luego `movieId`
- Matriz dispersa: no todos los usuarios califican todas las películas

### movies.csv

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `movieId` | int | Identificador de película (coincide con ratings.csv) |
| `title` | string | Título con año entre paréntesis (UTF-8), ej: "Toy Story (1995)" |
| `genres` | string | Géneros separados por `|`, ej: "Adventure\|Animation\|Comedy" |

**Géneros disponibles** (20): Action, Adventure, Animation, Children, Comedy, Crime, Documentary, Drama, Fantasy, Film-Noir, Horror, IMAX, Musical, Mystery, Romance, Sci-Fi, Thriller, War, Western, (no genres listed)

### links.csv

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `movieId` | int | Identificador de película |
| `imdbId` | string | ID de IMDB (7 dígitos con ceros a la izquierda) |
| `tmdbId` | int | ID de The Movie Database (TMDB) |

### tags.csv

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `userId` | int | Usuario que asignó el tag |
| `movieId` | int | Película etiquetada |
| `tag` | string | Etiqueta en texto libre (ej: "pixar", "visually appealing") |
| `timestamp` | int64 | Momento de asignación (UNIX timestamp) |

**Notas**: 
- Requiere normalización (lowercase, trim, deduplicación)
- Contiene typos y variantes ("pixar" vs "Pixar" vs "PIXAR")

### genome-tags.csv

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `tagId` | int | ID numérico del tag (1 a 1,128) |
| `tag` | string | Etiqueta curada del sistema Genome |

**Ejemplos**: "dystopia", "ensemble cast", "computer animation"

### genome-scores.csv

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `movieId` | int | Película evaluada |
| `tagId` | int | ID del genome tag |
| `relevance` | float | Relevancia del tag para la película (0.0 a 1.0) |

**Notas**:
- Scores generados algorítmicamente por MovieLens
- ~13.8M entradas (no todas las películas tienen todos los tags)
- Valores altos (>0.9) indican fuerte asociación

### item_map.csv (generado en PC3)

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `movieId` | int | ID original de MovieLens |
| `iIdx` | int | Índice remapeado continuo (0 a 32,719) |

**Propósito**: Mapear IDs dispersos a índices contiguos para matrices del modelo

### user_map.csv (generado en PC3)

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `userId` | int | ID original de MovieLens |
| `uIdx` | int | Índice remapeado continuo (0 a 162,540) |

**Propósito**: Mapear usuarios a índices contiguos para vectores del modelo

### item_topk_cosine_conc.csv (generado en PC3)

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `iIdx` | int | Índice de la película objetivo |
| `neighborIdx` | int | Índice del vecino similar |
| `similarity` | float | Similitud coseno (0.0 a 1.0) |

**Notas**:
- k=20 vecinos más similares por película
- Ordenado por similitud descendente
- ~600K filas (30K películas × 20 vecinos)

---

## 🔗 Relaciones Conceptuales

```
┌──────────────┐
│   ratings    │──────┐
│ userId       │      │
│ movieId  ────┼──┐   │
│ rating       │  │   │
└──────────────┘  │   │
                  │   │
        ┌─────────▼───▼──────┐        ┌──────────────┐
        │       movies        │◄───────│    links     │
        │ movieId (PK)        │        │ movieId      │
        │ title               │        │ imdbId       │
        │ genres          ────┼───┐    │ tmdbId       │
        └─────────────────────┘   │    └──────────────┘
                  │               │
                  │               │    ┌──────────────┐
        ┌─────────▼───────┐       └───►│     tags     │
        │  genome-scores  │            │ userId       │
        │ movieId         │            │ movieId      │
        │ tagId       ────┼───┐        │ tag          │
        │ relevance       │   │        └──────────────┘
        └─────────────────┘   │
                              │
                    ┌─────────▼──────┐
                    │  genome-tags   │
                    │ tagId (PK)     │
                    │ tag            │
                    └────────────────┘

┌──────────────┐        ┌──────────────────┐
│  user_map    │        │   item_map       │
│ userId   ────┼───┐    │ movieId      ────┼───┐
│ uIdx         │   │    │ iIdx             │   │
└──────────────┘   │    └──────────────────┘   │
                   │                            │
         ┌─────────▼────────────────────────────▼──────┐
         │     item_topk_cosine_conc (similitudes)     │
         │ iIdx (película)                             │
         │ neighborIdx (vecino similar)                │
         │ similarity (coseno)                         │
         └─────────────────────────────────────────────┘
```

**Flujo de datos**:
1. `ratings` conecta usuarios con películas mediante ratings
2. `movies` define el catálogo con metadatos básicos
3. `links` permite vincular con APIs externas (TMDB)
4. `tags` y `genome-scores` enriquecen películas con características
5. `item_map` y `user_map` mapean IDs originales a índices del modelo
6. `item_topk_cosine_conc` pre-calcula similitudes para recomendaciones

---

## 🎯 Relevancia del Proyecto

### Contexto del Sistema de Recomendación

Este ETL es un componente crítico dentro de un **sistema de recomendación de películas** que utiliza algoritmos de filtrado colaborativo y basado en contenido. El proyecto forma parte de una arquitectura completa que incluye:

1. **Backend (API REST)**: Consume los datos procesados por este ETL
2. **Frontend (Web/Mobile)**: Interfaz de usuario para navegación y recomendaciones
3. **Motor de Recomendaciones**: Utiliza las similitudes calculadas para sugerir contenido
4. **Base de Datos MongoDB**: Almacena toda la información enriquecida

### Problema que Resuelve

Los sistemas de recomendación modernos requieren:

- **Datos enriquecidos**: No basta con tener ratings; necesitamos metadatos (géneros, tags, sinopsis, cast)
- **Normalización**: Los datos crudos tienen inconsistencias (typos en tags, formatos diversos)
- **Integración externa**: APIs como TMDB proveen información visual y descriptiva esencial
- **Eficiencia**: Procesamiento de millones de registros (25M+ ratings, 162K+ usuarios)
- **Mapeo de IDs**: El sistema de recomendación usa índices remapeados (iIdx, uIdx) para optimización
- **Mapeo dinámico**: Asignación automática de índices a nuevos usuarios/películas

### Impacto en el Sistema

El ETL transforma datos crudos dispersos en **colecciones estructuradas** que permiten:

✅ **Recomendaciones precisas**: Similitud coseno pre-calculada (k=20 vecinos)  
✅ **Búsqueda enriquecida**: Tags normalizados y ordenados por popularidad  
✅ **Experiencia visual**: Posters y fotos del cast desde TMDB  
✅ **Análisis de usuarios**: Estadísticas de ratings y preferencias  
✅ **Escalabilidad**: Formato NDJSON optimizado para carga masiva en MongoDB  

---

## 🏗️ Arquitectura del Sistema

### Pipeline de Procesamiento

```
┌─────────────────────────────────────────────────────────────────────┐
│                         FASE 1: LOCAL DATA                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  CSV Sources:                    Processing:                         │
│  ┌──────────────┐               ┌──────────────┐                    │
│  │ movies.csv   │──────────────>│ Parse & Clean│                    │
│  │ ratings.csv  │──────────────>│ Normalize    │                    │
│  │ links.csv    │──────────────>│ Aggregate    │                    │
│  │ tags.csv     │──────────────>│ Deduplicate  │                    │
│  │ genome-*.csv │──────────────>│ Sort & Rank  │                    │
│  │ item_map.csv │──────────────>│ Map IDs      │                    │
│  │ user_map.csv │──────────────>│ Generate     │                    │
│  └──────────────┘               └──────┬───────┘                    │
│                                        │                              │
└────────────────────────────────────────┼──────────────────────────────┘
                                         │
┌────────────────────────────────────────┼──────────────────────────────┐
│                         FASE 2: EXTERNAL API                          │
├────────────────────────────────────────┼──────────────────────────────┤
│                                        ▼                              │
│                            ┌────────────────────┐                     │
│                            │ TMDB API Client    │                     │
│                            │ • Rate Limiting    │                     │
│                            │ • Caching          │                     │
│                            │ • Error Handling   │                     │
│                            └─────────┬──────────┘                     │
│                                      │                                │
│                            ┌─────────▼──────────┐                     │
│                            │ External Data:     │                     │
│                            │ • Posters          │                     │
│                            │ • Overview         │                     │
│                            │ • Cast + Photos    │                     │
│                            │ • Director         │                     │
│                            │ • Budget/Revenue   │                     │
│                            └─────────┬──────────┘                     │
└──────────────────────────────────────┼──────────────────────────────┘
                                       │
┌──────────────────────────────────────┼──────────────────────────────┐
│                         OUTPUT: NDJSON FILES                          │
├──────────────────────────────────────┼──────────────────────────────┤
│                                      ▼                                │
│  ┌───────────────────────────────────────────────────────────┐       │
│  │ movies.ndjson (62K docs)                                  │       │
│  │ • movieId, iIdx, title, year, genres                      │       │
│  │ • links (MovieLens, IMDB, TMDB)                           │       │
│  │ • genomeTags (top 10 by relevance)                        │       │
│  │ • userTags (top 10 by frequency)                          │       │
│  │ • ratingStats (avg, count, lastRatedAt)                   │       │
│  │ • externalData (TMDB: poster, cast, overview, etc.)       │       │
│  └───────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────┐       │
│  │ ratings.ndjson (25M docs)                                 │       │
│  │ • userId, movieId, rating, timestamp                      │       │
│  └───────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────┐       │
│  │ users.ndjson (162K docs)                                  │       │
│  │ • userId, uIdx, firstName, lastName, username             │       │
│  │ • email, passwordHash, role                               │       │
│  │ • about, preferredGenres[], createdAt, updatedAt          │       │
│  └───────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────┐       │
│  │ similarities.ndjson (30K docs)                            │       │
│  │ • _id: "{iIdx}_cosine_k20"                                │       │
│  │ • iIdx, movieId, metric: "cosine", k: 20                  │       │
│  │ • neighbors[] (movieId, iIdx, sim)                        │       │
│  └───────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────┐       │
│  │ passwords_log.csv (162K users)                            │       │
│  │ • userId, uIdx, firstName, lastName, username             │       │
│  │ • email, password, passwordHash                           │       │
│  └───────────────────────────────────────────────────────────┘       │
└───────────────────────────────────────────────────────────────────────┘
```

### Flujo de Integración con el Backend

```
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│   ETL (Go)   │─────>│   MongoDB    │<─────│Backend (API) │
│              │ load │              │query │              │
│ • Transform  │      │ • movies     │      │ • REST API   │
│ • Enrich     │      │ • ratings    │      │ • Auth       │
│ • Validate   │      │ • users      │      │ • Search     │
│ • Generate   │      │ • similarit. │      │ • Recommend  │
└──────────────┘      └──────────────┘      └──────┬───────┘
                                                    │
                                            ┌───────▼───────┐
                                            │   Frontend    │
                                            │               │
                                            │ • Web UI      │
                                            │ • Movie Cards │
                                            │ • Recommend.  │
                                            └───────────────┘
```

---

## 📚 Fundamentos Teóricos

### 1. ETL (Extract, Transform, Load)

El proceso ETL es fundamental en la ingeniería de datos:

- **Extract**: Lectura de múltiples fuentes CSV (7 archivos principales)
- **Transform**: 
  - Normalización de texto (lowercase, trim, deduplicación)
  - Agregación de estadísticas (ratings promedio)
  - Ranking por relevancia/frecuencia (genome tags, user tags)
  - Mapeo de IDs (movieId↔iIdx, userId↔uIdx)
  - **Mapeo dinámico**: Asignación automática de índices a nuevas entidades
  - Integración de APIs externas (TMDB)
- **Load**: Generación de NDJSON para importación masiva en MongoDB

### 2. Similitud de Items (Cosine Similarity)

El archivo `item_topk_cosine_conc.csv` contiene similitudes pre-calculadas usando la fórmula:

**sim(i, j) = cos(θ) = (A · B) / (||A|| × ||B||)**

Donde:
- **i, j**: Películas representadas por vectores de ratings
- **sim(i,j)**: Similitud entre 0 y 1 (1 = idénticas)
- **k=20**: Top 20 vecinos más similares

**Aplicación**: Recomendaciones del tipo "Si te gustó X, también te puede gustar Y"

### 3. Genome Tags vs User Tags

#### Genome Tags
- **Origen**: Sistema algorítmico de MovieLens
- **Características**: 1,128 tags predefinidos con scores de relevancia (0.0-1.0)
- **Ejemplo**: "pixar animation" (0.9957), "computer animation" (0.9987)
- **Uso**: Búsqueda y filtrado por características específicas

#### User Tags
- **Origen**: Etiquetas manuales de usuarios
- **Características**: Texto libre, requiere normalización
- **Procesamiento**: 
  - Lowercase + trim
  - Deduplicación
  - Ranking por frecuencia (cuántos usuarios asignaron ese tag)
- **Ejemplo**: "pixar" (asignado por 150 usuarios) > "nice movie" (2 usuarios)
- **Uso**: Descubrimiento de tendencias y preferencias de la comunidad

### 4. Mapeo de IDs (Remapping)

**Problema**: Los IDs originales (movieId, userId) tienen gaps y no son secuenciales.

**Solución**: 
- `item_map.csv`: movieId → iIdx (0 a N-1 continuo)
- `user_map.csv`: userId → uIdx (0 a M-1 continuo)

**Mapeo Dinámico**:
- `IDMapper`: Estructura thread-safe con `sync.RWMutex`
- `GetOrCreate(id)`: Asigna automáticamente el siguiente índice disponible a IDs nuevos
- `--update-mappings`: Flag para persistir cambios a CSVs

**Beneficio**: 
- Algoritmos de recomendación operan con matrices densas
- Reducción de memoria (índices contiguos)
- Soporte automático para nuevas películas/usuarios sin regenerar modelo completo

### 5. Rate Limiting y Caching (TMDB API)

**Rate Limiting**:
```go
rateLimiter: time.Tick(time.Second / 4) // 4 req/s
<-rateLimiter // Espera antes de cada request
```

**Caching in-memory**:
- Evita duplicados en una misma ejecución
- Reduce llamadas API (costo/latencia)
- Thread-safe con `sync.RWMutex`

**TMDB Limits**: 40 requests cada 10 segundos ≈ 4 req/s

### 6. Hashing de Passwords (bcrypt)

**Algoritmo bcrypt**:
- Cost factor: 10 (2^10 = 1024 iteraciones)
- Salting automático (previene rainbow tables)
- Resistente a ataques de fuerza bruta

**Trade-off**:
- `--hash-passwords=true`: Seguro pero lento (~162K usuarios en ~10 min)
- `--hash-passwords=false`: Rápido pero inseguro (~162K usuarios en ~5 seg)

**Recomendación**: Usar `false` en desarrollo, `true` en producción.

### 7. Generación de Perfiles de Usuario

**Sistema de generación automática de datos realistas**:

El ETL genera perfiles completos para los 162K+ usuarios utilizando datos aleatorios pero coherentes:

**Generación de Nombres**:
- Librería: `github.com/jaswdr/faker`
- Método: `GenerateRandomName()` produce combinaciones únicas de nombres y apellidos
- Ejemplo: "Alexander Johnson", "Sophia Williams", "Michael Chen"

**Generación de Usernames**:
- Formato: `firstname.lastname{número}`
- Número basado en `userId % 10000` para garantizar unicidad
- Ejemplo: "alexander.johnson123", "sophia.williams4567"

**Generación de About (Biografía)**:
- **70% con géneros**: Templates que mencionan géneros favoritos
  - "Fan of Action and Drama movies"
  - "Passionate about Comedy cinema"
  - "I really like Thriller films"
- **30% frases simples**: Descripciones genéricas
  - "Movie lover"
  - "Film enthusiast"
  - "Cinema addict"

**Selección de Géneros Preferidos**:
- Extracción: Se leen todos los géneros únicos de `movies.csv` al inicio
- Cantidad: Entre 1 y 5 géneros aleatorios por usuario
- Método: Fisher-Yates shuffle para selección aleatoria uniforme
- Los géneros se almacenan como array de strings para facilitar queries en MongoDB

**Timestamps**:
- `createdAt`: Timestamp ISO 8601 del momento de generación
- `updatedAt`: Inicialmente igual a `createdAt` (se actualizará en futuras modificaciones)

**Trazabilidad**:
- `passwords_log.csv` incluye todos los campos generados (firstName, lastName, username)
- Permite auditoría y debugging del proceso de generación

---

## 📦 Colecciones Generadas

### 1. `movies` (62,423 documentos)

```json
{
  "movieId": 1,
  "iIdx": 70,
  "title": "Toy Story",
  "year": 1995,
  "genres": ["Adventure", "Animation", "Children", "Comedy", "Fantasy"],
  "links": {
    "movielens": "https://movielens.org/movies/1",
    "imdb": "http://www.imdb.com/title/tt0114709/",
    "tmdb": "https://www.themoviedb.org/movie/862"
  },
  "genomeTags": [
    {"tag": "toys", "relevance": 0.99925},
    {"tag": "computer animation", "relevance": 0.99875}
  ],
  "userTags": [
    "pixar", "animation", "disney", "tom hanks", "computer animation"
  ],
  "ratingStats": {
    "average": 3.89,
    "count": 57309,
    "lastRatedAt": "2019-11-20T21:23:42Z"
  },
  "externalData": {
    "posterUrl": "https://image.tmdb.org/t/p/w500/...",
    "overview": "Led by Woody, Andy's toys live happily...",
    "cast": [
      {
        "name": "Tom Hanks",
        "profileUrl": "https://image.tmdb.org/t/p/w185/..."
      }
    ],
    "director": "John Lasseter",
    "runtime": 81,
    "budget": 30000000,
    "revenue": 394436586,
    "tmdbFetched": true
  },
  "createdAt": "2025-11-21T20:33:00Z",
  "updatedAt": "2025-11-21T20:33:00Z"
}
```

**Características**:
- ✅ **iIdx**: ID remapeado para el modelo de recomendación
- ✅ **genomeTags**: Top 10 por relevancia (≥0.5)
- ✅ **userTags**: Top 10 por frecuencia (normalizados)
- ✅ **externalData**: Cast con fotos de perfil

### 2. `ratings` (25,000,095 documentos)

```json
{
  "userId": 1,
  "movieId": 296,
  "rating": 5.0,
  "timestamp": 1147880044
}
```

**Uso**: Entrenar modelos de filtrado colaborativo

### 3. `users` (162,541 documentos)

```json
{
  "userId": 1,
  "uIdx": 0,
  "firstName": "Alexander",
  "lastName": "Johnson",
  "username": "alexander.johnson1",
  "email": "user1@email.com",
  "passwordHash": "$2a$10$...",
  "role": "user",
  "about": "Fan of Action and Drama movies",
  "preferredGenres": ["Action", "Drama", "Thriller"],
  "createdAt": "2025-11-27T10:15:30Z",
  "updatedAt": "2025-11-27T10:15:30Z"
}
```

**Características**:
- ✅ **uIdx**: ID remapeado para el modelo
- ✅ **firstName, lastName**: Nombres aleatorios generados con [faker](https://github.com/jaswdr/faker)
- ✅ **username**: Formato `firstname.lastname` + número único
- ✅ **email**: Generado automáticamente como `user{userId}@email.com`
- ✅ **passwordHash**: bcrypt (opcional con `--hash-passwords`)
- ✅ **about**: Descripción personalizada (70% menciona géneros favoritos, 30% frases simples)
- ✅ **preferredGenres**: Array de 1-5 géneros aleatorios extraídos de movies.csv
- ✅ **updatedAt**: Igual a createdAt inicialmente
- ✅ **Log disponible**: `passwords_log.csv` con passwords sin hashear y datos completos del perfil

### 4. `similarities` (30,202 documentos)

```json
{
  "_id": "16490_cosine_k20",
  "movieId": 26010,
  "iIdx": 16490,
  "metric": "cosine",
  "k": 20,
  "neighbors": [
    {"movieId": 69908, "iIdx": 21813, "sim": 0.140301},
    {"movieId": 31297, "iIdx": 21720, "sim": 0.108906}
  ],
  "updatedAt": "2025-11-21T22:39:34Z"
}
```

**Uso**: Sistema de recomendaciones basado en similitud de items

---

## 🛠️ Tecnologías Utilizadas

### Lenguaje y Librerías

- **Go 1.21+**: Eficiencia, concurrencia nativa, bajo consumo de memoria
- **Librerías estándar**: `encoding/csv`, `encoding/json`, `net/http`, `bufio`, `sync`
- **bcrypt**: `golang.org/x/crypto/bcrypt` para hashing de passwords
- **faker**: `github.com/jaswdr/faker` para generación de datos aleatorios de usuarios (nombres, apellidos)
- **IDMapper**: Sistema de mapeo dinámico thread-safe para gestión de índices

### Base de Datos

- **MongoDB 4.4+**: Base de datos NoSQL orientada a documentos
- **NDJSON**: Formato optimizado para importación masiva (`mongoimport`)

### APIs Externas

- **TMDB API v3**: The Movie Database
  - Endpoint: `https://api.themoviedb.org/3/`
  - Rate limit: 40 req/10s
  - Documentación: https://developers.themoviedb.org/3

### Dataset

- **MovieLens 25M**: 
  - 25M ratings
  - 62K películas
  - 162K usuarios
  - Fuente: https://grouplens.org/datasets/movielens/

---

## 📊 Estadísticas del Procesamiento

### Dataset Procesado

| Colección | Registros | Tamaño (NDJSON) | Tiempo Estimado |
|-----------|-----------|-----------------|-----------------|
| movies | 62,423 | ~150 MB | 2 min (sin API) / 5 horas (con API) |
| ratings | 25,000,095 | ~1.5 GB | 3 min |
| users | 162,541 | ~25 MB | 5 seg (sin hash) / 10 min (con hash) |
| similarities | 30,202 | ~50 MB | 1 min |

### Tiempos de Ejecución (hardware promedio)

- **Fase 1 (solo local)**: ~5-7 minutos
- **Fase 2 (con TMDB API)**: ~4-5 horas (debido a rate limiting)
- **Prueba rápida** (`movies_test.csv`, 10 películas): ~10 segundos

---

## 🎓 Casos de Uso

### 1. Sistema de Recomendación
- **Content-Based**: Usar genomeTags para recomendar películas similares por características
- **Collaborative Filtering**: Usar ratings para recomendar basado en usuarios similares
- **Hybrid**: Combinar similitudes pre-calculadas con ratings en tiempo real

### 2. Búsqueda Avanzada
- **Por tags**: Buscar "disney animation" usando genomeTags
- **Por popularidad**: Ordenar por ratingStats.count
- **Por género**: Filtrar por genres array

### 3. Análisis de Datos
- **Tendencias**: Analizar userTags más frecuentes por año
- **Taquilla**: Correlacionar budget vs revenue (TMDB)
- **Engagement**: Identificar películas con más ratings recientes

### 4. Interfaz de Usuario
- **Movie Cards**: Mostrar poster, título, rating promedio
- **Cast Grid**: Fotos del elenco con nombres
- **Similar Movies**: Top 5 vecinos de similarities

---

## 📖 Guía de Uso

Para instrucciones detalladas de instalación, configuración y ejecución, consulta **[GUIDE.md](./GUIDE.md)**.

---

## 📄 Licencia

Este proyecto es parte de un trabajo académico para el curso de Programación Concurrente y Distribuida - UPC 2025.

Dataset: MovieLens 25M © GroupLens Research  
TMDB Data: © The Movie Database (TMDb)

---

## 👥 Autores

**Grupo 3**

Proyecto desarrollado como parte del curso de **Programación Concurrente y Distribuida** - Universidad Peruana de Ciencias Aplicadas (UPC), ciclo 2025-2.

### Integrantes

| Nombre | Código de Estudiante |
|--------|----------------------|
| Marsi Valeria Figueroa Larragán | U202220990 |
| Liam Mikael Quino Neff | U20221E167 |
| Mauricio Eduardo Vera Castellón | U20181H114 |

**Repositorio**: [PrograCyD/PC4_ETLConstructionWithMongoDB](https://github.com/PrograCyD/PC4_ETLConstructionWithMongoDB)

---

## 📚 Referencias

1. Harper, F. M., & Konstan, J. A. (2015). The MovieLens Datasets: History and Context. ACM Transactions on Interactive Intelligent Systems.
2. The Movie Database (TMDb) API Documentation: https://developers.themoviedb.org/3
3. MongoDB Manual: https://docs.mongodb.com/manual/
4. Go Programming Language: https://go.dev/doc/
5. bcrypt Paper: Provos, N., & Mazières, D. (1999). A Future-Adaptable Password Scheme.
