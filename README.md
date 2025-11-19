# Importar MovieLens a MongoDB

Este repositorio contiene una herramienta escrita en Go que convierte los CSV de MovieLens
(`movies.csv` y `ratings.csv`) a NDJSON listos para importar en MongoDB.

Resumen rápido
- `convert_to_mongo_json.go`: programa Go que lee `movies.csv` y `ratings.csv` en streaming
  y genera `out/movies.ndjson`, `out/ratings.ndjson` y opcionalmente `out/users.ndjson`.
- Los datasets provienen del conjunto MovieLens (archivos `movies.csv` y `ratings.csv`).

Cómo funciona (Go)
- El programa parsea cada fila y genera documentos JSON por línea (NDJSON):
  - `movies.ndjson`: { movieId, title, year?, genres[], createdAt }
  - `ratings.ndjson`: { userId, movieId, rating, timestamp }
  - `users.ndjson` (opcional): { userId, createdAt } — generado agregando todos los `userId` encontrados en `ratings`.
- Extrae el `year` del título cuando el título contiene `(YYYY)`.
- Convierte `genres` separados por `|` en arrays de strings.
- Es un procesador en streaming: puede manejar ficheros grandes eficientemente.

Compilar y ejecutar (PowerShell)

1) Compilar el binario (opcional):

```powershell
go build -o .\out\convert_to_mongo_json.exe .\convert_to_mongo_json.go
```

2) Ejecutar con `go run` o usando el binario. Genera `movies.ndjson` y `ratings.ndjson`:

```powershell
# usando go run
go run .\convert_to_mongo_json.go --data-dir . --out-dir .\out

# o usando el binario generado
.\out\convert_to_mongo_json.exe --data-dir . --out-dir .\out
```

3) Para generar también `users.ndjson` (crea en memoria el conjunto de userId):

```powershell
go run .\convert_to_mongo_json.go --data-dir . --out-dir .\out --generate-users
```

Salida
- Los ficheros NDJSON se escriben en el directorio indicado con `--out-dir` (por defecto `out`).
- Cada línea es un documento JSON independiente, listo para `mongoimport`.

Importar a MongoDB
Reemplaza la `--uri` por la URI de tu cluster/local:

```powershell
mongoimport --uri "mongodb+srv://<user>:<pass>@cluster0.mongodb.net/mydb" --collection movies --file .\out\movies.ndjson --type json
mongoimport --uri "mongodb+srv://<user>:<pass>@cluster0.mongodb.net/mydb" --collection ratings --file .\out\ratings.ndjson --type json
mongoimport --uri "mongodb+srv://<user>:<pass>@cluster0.mongodb.net/mydb" --collection users --file .\out\users.ndjson --type json
```

Crear `users` directamente en Mongo (alternativa sin generar `users.ndjson` localmente)

```js
db.ratings.aggregate([
  { $group: { _id: "$userId" } },
  { $project: { userId: "$_id", createdAt: new Date() } },
  { $merge: { into: "users", on: "userId", whenMatched: "keepExisting", whenNotMatched: "insert" } }
])
```

Índices recomendados (Ejemplos)

```js
db.users.createIndex({ userId: 1 }, { unique: true })
db.movies.createIndex({ movieId: 1 }, { unique: true })
db.movies.createIndex({ title: "text" })
db.ratings.createIndex({ userId: 1 })
db.ratings.createIndex({ movieId: 1 })
```

Notas importantes
- Los CSV (`movies.csv`, `ratings.csv`) son del conjunto MovieLens; no incluyen `username` ni `email`.
- `--generate-users` mantiene un conjunto de `userId` en memoria: para conjuntos muy grandes puede consumir RAM,
  aunque el consumo suele ser razonable para decenas de millones de entradas.
- `similarities` y `recommendations` no se generan aquí: requieren ejecutar un motor de recomendación (algoritmos Pearson/Cosine, etc.)
  sobre la colección `ratings` y luego guardar los resultados en las colecciones `similarities` y `recommendations`.

Si quieres, puedo añadir un ejemplo (notebook o script) para calcular similitudes y generar la colección `similarities`.
# Importar MovieLens a MongoDB

Archivos generados por el script `convert_to_mongo_json.py`:
- `movies.ndjson` → colección `movies`
- `ratings.ndjson` → colección `ratings`
- `users.ndjson` (opcional) → colección `users` (si se generó con `--generate-users`)

Ejemplo: generar los NDJSON

```powershell
python .\convert_to_mongo_json.py --data-dir . --out-dir .\out --generate-users
```

Comandos `mongoimport` (ejemplos):

1) Importar `movies` (ndjson, cada línea es un documento):

```powershell
mongoimport --uri "mongodb+srv://<user>:<pass>@cluster0.mongodb.net/mydb" --collection movies --file .\out\movies.ndjson --type json
```

2) Importar `ratings`:

```powershell
mongoimport --uri "mongodb+srv://<user>:<pass>@cluster0.mongodb.net/mydb" --collection ratings --file .\out\ratings.ndjson --type json
```

3) Importar `users` (si lo generaste):

```powershell
mongoimport --uri "mongodb+srv://<user>:<pass>@cluster0.mongodb.net/mydb" --collection users --file .\out\users.ndjson --type json
```

Si prefieres no generar `users.ndjson` localmente (evita usar mucha memoria), puedes crear la colección `users` directamente en Mongo a partir de `ratings` con una agregación:

```js
// En Mongo shell o en Compass - Aggregation
db.ratings.aggregate([
  { $group: { _id: "$userId" } },
  { $project: { userId: "$_id", createdAt: new Date() } },
  { $merge: { into: "users", on: "userId", whenMatched: "keepExisting", whenNotMatched: "insert" } }
])
```

Índices recomendados (ejemplos):

```js
db.users.createIndex({ userId: 1 }, { unique: true })
db.movies.createIndex({ movieId: 1 }, { unique: true })
db.movies.createIndex({ title: "text" })
db.ratings.createIndex({ userId: 1 })
db.ratings.createIndex({ movieId: 1 })

// Para similarities / recommendations (después de calcularlos):
db.similarities.createIndex({ userId: 1, similarityMetric: 1 })
db.recommendations.createIndex({ userId: 1, algo: 1, similarityMetric: 1 })
```

Qué se puede poblar automáticamente desde los CSV que tienes
- `movies`: sí — contiene `movieId`, `title`, `genres` y se extrae `year` del título si existe.
- `ratings`: sí — contiene `userId`, `movieId`, `rating`, `timestamp`.
- `users`: puede generarse desde `ratings` (solo `userId` y `createdAt`), pero no hay `username` ni `email` en los CSV — esos campos tendrían que añadirse manualmente o por importación externa.

Qué falta y debe calcularse/manual:
- `similarities`: requiere ejecutar el algoritmo de similitud (Pearson / Cosine) sobre `ratings` — no se llena desde CSV.
- `recommendations`: requiere ejecutar el motor de recomendación (user-based o item-based) y guardar los resultados.
