package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var yearRe = regexp.MustCompile(`\((\d{4})\)\s*$`)

func isoNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func parseTitleAndYear(raw string) (string, *int) {
	raw = strings.TrimSpace(raw)
	m := yearRe.FindStringSubmatch(raw)
	if len(m) == 2 {
		y, err := strconv.Atoi(m[1])
		if err == nil {
			// remove last occurrence of (YYYY)
			idx := strings.LastIndex(raw, "(")
			if idx > 0 {
				title := strings.TrimSpace(raw[:idx])
				return title, &y
			}
			return strings.TrimSpace(raw), &y
		}
	}
	// fallback: no year
	return raw, nil
}

type MovieDoc struct {
	MovieID   int      `json:"movieId"`
	Title     string   `json:"title"`
	Year      *int     `json:"year,omitempty"`
	Genres    []string `json:"genres"`
	CreatedAt string   `json:"createdAt"`
}

type RatingDoc struct {
	UserID    int     `json:"userId"`
	MovieID   int     `json:"movieId"`
	Rating    float64 `json:"rating"`
	Timestamp int64   `json:"timestamp"`
}

type UserDoc struct {
	UserID    int    `json:"userId"`
	CreatedAt string `json:"createdAt"`
}

func processMovies(inPath, outPath string) (int, error) {
	f, err := os.Open(inPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	r := csv.NewReader(bufio.NewReader(f))
	r.FieldsPerRecord = -1

	// open output
	of, err := os.Create(outPath)
	if err != nil {
		return 0, err
	}
	defer of.Close()
	w := bufio.NewWriter(of)
	defer w.Flush()

	// read header
	header, err := r.Read()
	if err != nil {
		return 0, err
	}
	idx := map[string]int{}
	for i, h := range header {
		idx[h] = i
	}

	written := 0
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// skip malformed
			continue
		}
		// guard indexes
		mid := 0
		if v, ok := idx["movieId"]; ok && v < len(rec) {
			mid, _ = strconv.Atoi(rec[v])
		} else if len(rec) > 0 {
			mid, _ = strconv.Atoi(rec[0])
		}
		titleRaw := ""
		if v, ok := idx["title"]; ok && v < len(rec) {
			titleRaw = rec[v]
		} else if len(rec) > 1 {
			titleRaw = rec[1]
		}
		genresRaw := ""
		if v, ok := idx["genres"]; ok && v < len(rec) {
			genresRaw = rec[v]
		} else if len(rec) > 2 {
			genresRaw = rec[2]
		}

		title, year := parseTitleAndYear(titleRaw)
		genres := []string{}
		if genresRaw != "" && genresRaw != "(no genres listed)" {
			for _, g := range strings.Split(genresRaw, "|") {
				g = strings.TrimSpace(g)
				if g != "" {
					genres = append(genres, g)
				}
			}
		}

		doc := MovieDoc{
			MovieID:   mid,
			Title:     title,
			Year:      year,
			Genres:    genres,
			CreatedAt: isoNow(),
		}
		b, _ := json.Marshal(doc)
		w.Write(b)
		w.WriteByte('\n')
		written++
	}
	return written, nil
}

func processRatings(inPath, outPath string, generateUsers bool) (int, error) {
	f, err := os.Open(inPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	r := csv.NewReader(bufio.NewReader(f))
	r.FieldsPerRecord = -1

	of, err := os.Create(outPath)
	if err != nil {
		return 0, err
	}
	defer of.Close()
	w := bufio.NewWriter(of)
	defer w.Flush()

	header, err := r.Read()
	if err != nil {
		return 0, err
	}
	idx := map[string]int{}
	for i, h := range header {
		idx[h] = i
	}

	users := map[int]struct{}{}
	written := 0
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		uid := 0
		mid := 0
		rating := 0.0
		ts := int64(0)
		if v, ok := idx["userId"]; ok && v < len(rec) {
			uid, _ = strconv.Atoi(rec[v])
		} else if len(rec) > 0 {
			uid, _ = strconv.Atoi(rec[0])
		}
		if v, ok := idx["movieId"]; ok && v < len(rec) {
			mid, _ = strconv.Atoi(rec[v])
		} else if len(rec) > 1 {
			mid, _ = strconv.Atoi(rec[1])
		}
		if v, ok := idx["rating"]; ok && v < len(rec) {
			rating, _ = strconv.ParseFloat(rec[v], 64)
		} else if len(rec) > 2 {
			rating, _ = strconv.ParseFloat(rec[2], 64)
		}
		if v, ok := idx["timestamp"]; ok && v < len(rec) {
			ts, _ = strconv.ParseInt(rec[v], 10, 64)
		} else if len(rec) > 3 {
			ts, _ = strconv.ParseInt(rec[3], 10, 64)
		}

		doc := RatingDoc{
			UserID:    uid,
			MovieID:   mid,
			Rating:    rating,
			Timestamp: ts,
		}
		b, _ := json.Marshal(doc)
		w.Write(b)
		w.WriteByte('\n')
		written++
		if generateUsers {
			users[uid] = struct{}{}
		}
	}

	if generateUsers {
		usersSlice := make([]int, 0, len(users))
		for k := range users {
			usersSlice = append(usersSlice, k)
		}
		sort.Ints(usersSlice)
		usersOut := strings.Replace(outPath, "ratings.ndjson", "users.ndjson", 1)
		if usersOut == outPath {
			usersOut = strings.TrimSuffix(outPath, filepath.Ext(outPath)) + "_users.ndjson"
		}
		uf, err := os.Create(usersOut)
		if err == nil {
			uw := bufio.NewWriter(uf)
			for _, uid := range usersSlice {
				ud := UserDoc{UserID: uid, CreatedAt: isoNow()}
				b, _ := json.Marshal(ud)
				uw.Write(b)
				uw.WriteByte('\n')
			}
			uw.Flush()
			uf.Close()
		}
	}

	return written, nil
}

func main() {
	dataDir := flag.String("data-dir", ".", "Directorio con los csv (default: .)")
	moviesFile := flag.String("movies-file", "movies.csv", "Nombre de movies.csv")
	ratingsFile := flag.String("ratings-file", "ratings.csv", "Nombre de ratings.csv")
	outDir := flag.String("out-dir", "out", "Directorio de salida para NDJSON")
	genUsers := flag.Bool("generate-users", false, "Generar users.ndjson desde ratings (usa memoria)")
	flag.Parse()

	os.MkdirAll(*outDir, 0o755)

	moviesPath := filepath.Join(*dataDir, *moviesFile)
	ratingsPath := filepath.Join(*dataDir, *ratingsFile)
	moviesOut := filepath.Join(*outDir, "movies.ndjson")
	ratingsOut := filepath.Join(*outDir, "ratings.ndjson")

	fmt.Println("Procesando movies:", moviesPath)
	mcount, merr := processMovies(moviesPath, moviesOut)
	if merr != nil {
		fmt.Fprintln(os.Stderr, "error procesando movies:", merr)
		os.Exit(1)
	}
	fmt.Printf("Escritas %d entradas en %s\n", mcount, moviesOut)

	fmt.Println("Procesando ratings:", ratingsPath)
	rcount, rerr := processRatings(ratingsPath, ratingsOut, *genUsers)
	if rerr != nil {
		fmt.Fprintln(os.Stderr, "error procesando ratings:", rerr)
		os.Exit(1)
	}
	fmt.Printf("Escritas %d entradas en %s\n", rcount, ratingsOut)

	if *genUsers {
		fmt.Println("Se generó users.ndjson junto al fichero de ratings.")
	} else {
		fmt.Println("No se generó users.ndjson. Si lo deseas, ejecuta con --generate-users")
	}
}
