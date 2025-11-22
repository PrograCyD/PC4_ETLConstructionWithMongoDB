package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// LoadEnvFile carga variables de entorno desde un archivo .env
func LoadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignorar líneas vacías y comentarios
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Parsear KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remover comillas si existen
			value = strings.Trim(value, "\"'")
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

// FormatDuration convierte una duración en un formato legible
func FormatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// GenerateReport genera un archivo de reporte con estadísticas del ETL
func GenerateReport(path string, moviesCount, ratingsCount, usersCount, similaritiesCount int, hashedPasswords, fetchedExternal bool, processedMovies, processedRatings, processedUsers, processedSimilarities bool, elapsed time.Duration) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	// Encabezado
	fmt.Fprintln(w, "================================================================================")
	fmt.Fprintln(w, "               ETL CONSTRUCTION WITH MONGODB - REPORTE DE EJECUCIÓN")
	fmt.Fprintln(w, "================================================================================")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Fecha de ejecución: %s\n", time.Now().Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(w, "Tiempo total: %s\n", FormatDuration(elapsed))
	fmt.Fprintln(w)

	// Configuración
	fmt.Fprintln(w, "CONFIGURACIÓN:")
	fmt.Fprintln(w, strings.Repeat("-", 80))
	if fetchedExternal {
		fmt.Fprintln(w, "  ✓ Fase 2: Datos enriquecidos con TMDB API")
	} else {
		fmt.Fprintln(w, "  • Fase 1: Solo datos locales (sin TMDB)")
	}
	if hashedPasswords {
		fmt.Fprintln(w, "  ✓ Passwords hasheados con bcrypt (seguro)")
	} else {
		fmt.Fprintln(w, "  ⚠ Passwords sin hashear (modo desarrollo)")
	}
	fmt.Fprintln(w)

	// Ejecución selectiva
	fmt.Fprintln(w, "PROCESADORES EJECUTADOS:")
	fmt.Fprintln(w, strings.Repeat("-", 80))
	if processedMovies {
		fmt.Fprintln(w, "  ✓ Movies")
	} else {
		fmt.Fprintln(w, "  ⏭ Movies (omitido)")
	}
	if processedRatings {
		fmt.Fprintln(w, "  ✓ Ratings")
	} else {
		fmt.Fprintln(w, "  ⏭ Ratings (omitido)")
	}
	if processedUsers {
		fmt.Fprintln(w, "  ✓ Users")
	} else {
		fmt.Fprintln(w, "  ⏭ Users (omitido)")
	}
	if processedSimilarities {
		fmt.Fprintln(w, "  ✓ Similarities")
	} else {
		fmt.Fprintln(w, "  ⏭ Similarities (omitido)")
	}
	fmt.Fprintln(w)

	// Estadísticas
	fmt.Fprintln(w, "ESTADÍSTICAS DE DATOS PROCESADOS:")
	fmt.Fprintln(w, strings.Repeat("-", 80))
	if processedMovies {
		fmt.Fprintf(w, "  Movies:        %10d documentos generados\n", moviesCount)
	} else {
		fmt.Fprintln(w, "  Movies:        (no procesado)")
	}
	if processedRatings {
		fmt.Fprintf(w, "  Ratings:       %10d documentos generados\n", ratingsCount)
	} else {
		fmt.Fprintln(w, "  Ratings:       (no procesado)")
	}
	if processedUsers {
		fmt.Fprintf(w, "  Users:         %10d documentos generados\n", usersCount)
	} else {
		fmt.Fprintln(w, "  Users:         (no procesado)")
	}
	if processedSimilarities {
		fmt.Fprintf(w, "  Similarities:  %10d documentos generados\n", similaritiesCount)
	} else {
		fmt.Fprintln(w, "  Similarities:  (no procesado)")
	}
	fmt.Fprintln(w, strings.Repeat("-", 80))
	total := 0
	if processedMovies {
		total += moviesCount
	}
	if processedRatings {
		total += ratingsCount
	}
	if processedUsers {
		total += usersCount
	}
	if processedSimilarities {
		total += similaritiesCount
	}
	fmt.Fprintf(w, "  TOTAL:         %10d documentos\n", total)
	fmt.Fprintln(w)

	// Archivos generados
	fmt.Fprintln(w, "ARCHIVOS GENERADOS:")
	fmt.Fprintln(w, strings.Repeat("-", 80))
	if processedMovies {
		fmt.Fprintln(w, "  • out/movies.ndjson         - Películas con metadata completa")
	}
	if processedRatings {
		fmt.Fprintln(w, "  • out/ratings.ndjson        - Valoraciones de usuarios")
	}
	if processedUsers {
		fmt.Fprintln(w, "  • out/users.ndjson          - Usuarios con credenciales")
		fmt.Fprintln(w, "  • out/passwords_log.csv     - Log de passwords (desarrollo)")
	}
	if processedSimilarities {
		fmt.Fprintln(w, "  • out/similarities.ndjson   - Similitudes coseno (k=20)")
	}
	fmt.Fprintln(w, "  • out/report.txt            - Este reporte")
	fmt.Fprintln(w)

	// Comandos de importación (solo para procesadores ejecutados)
	fmt.Fprintln(w, "IMPORTACIÓN A MONGODB:")
	fmt.Fprintln(w, strings.Repeat("-", 80))
	fmt.Fprintln(w, "Ejecutar los siguientes comandos en PowerShell:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  $DB = \"movielens\"")
	fmt.Fprintln(w, "  $OUT_DIR = \"out\"")
	fmt.Fprintln(w)
	if processedMovies {
		fmt.Fprintln(w, "  mongoimport --db $DB --collection movies --file \"$OUT_DIR\\movies.ndjson\"")
	}
	if processedRatings {
		fmt.Fprintln(w, "  mongoimport --db $DB --collection ratings --file \"$OUT_DIR\\ratings.ndjson\"")
	}
	if processedUsers {
		fmt.Fprintln(w, "  mongoimport --db $DB --collection users --file \"$OUT_DIR\\users.ndjson\"")
	}
	if processedSimilarities {
		fmt.Fprintln(w, "  mongoimport --db $DB --collection similarities --file \"$OUT_DIR\\similarities.ndjson\"")
	}
	fmt.Fprintln(w)

	// Verificación
	fmt.Fprintln(w, "VERIFICACIÓN EN MONGODB:")
	fmt.Fprintln(w, strings.Repeat("-", 80))
	fmt.Fprintln(w, "Ejecutar en mongosh para verificar:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  use movielens")
	if processedMovies {
		fmt.Fprintf(w, "  db.movies.countDocuments()       // Esperado: %d\n", moviesCount)
	}
	if processedRatings {
		fmt.Fprintf(w, "  db.ratings.countDocuments()      // Esperado: %d\n", ratingsCount)
	}
	if processedUsers {
		fmt.Fprintf(w, "  db.users.countDocuments()        // Esperado: %d\n", usersCount)
	}
	if processedSimilarities {
		fmt.Fprintf(w, "  db.similarities.countDocuments() // Esperado: %d\n", similaritiesCount)
	}
	fmt.Fprintln(w)

	// Índices recomendados
	fmt.Fprintln(w, "ÍNDICES RECOMENDADOS:")
	fmt.Fprintln(w, strings.Repeat("-", 80))
	if processedMovies {
		fmt.Fprintln(w, "  db.movies.createIndex({ movieId: 1 })")
		fmt.Fprintln(w, "  db.movies.createIndex({ iIdx: 1 })")
		fmt.Fprintln(w, "  db.movies.createIndex({ title: \"text\" })")
	}
	if processedRatings {
		fmt.Fprintln(w, "  db.ratings.createIndex({ userId: 1, movieId: 1 })")
	}
	if processedUsers {
		fmt.Fprintln(w, "  db.users.createIndex({ userId: 1 }, { unique: true })")
		fmt.Fprintln(w, "  db.users.createIndex({ email: 1 }, { unique: true })")
	}
	if processedSimilarities {
		fmt.Fprintln(w, "  db.similarities.createIndex({ iIdx: 1 })")
	}
	fmt.Fprintln(w)

	// Notas finales
	fmt.Fprintln(w, "NOTAS:")
	fmt.Fprintln(w, strings.Repeat("-", 80))
	fmt.Fprintln(w, "  • Los IDs (iIdx, uIdx) son mapeos para optimización de modelos ML")
	if processedMovies {
		fmt.Fprintln(w, "  • GenomeTags limitados a top 10 por relevancia (>= 0.5)")
		fmt.Fprintln(w, "  • UserTags limitados a top 10 por frecuencia de uso")
	}
	if fetchedExternal && processedMovies {
		fmt.Fprintln(w, "  • Cast incluye profileUrl de TMDB (w185)")
		fmt.Fprintln(w, "  • Posters, sinopsis y runtime obtenidos de TMDB")
	}
	if !hashedPasswords && processedUsers {
		fmt.Fprintln(w, "  ⚠ IMPORTANTE: Passwords sin hashear - NO usar en producción")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Para más información consultar README.md y GUIDE.md")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "================================================================================")
	fmt.Fprintln(w, "                          ¡ETL COMPLETADO EXITOSAMENTE!")
	fmt.Fprintln(w, "================================================================================")

	return nil
}
