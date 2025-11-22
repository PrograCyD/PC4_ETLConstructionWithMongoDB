package mappers

import (
	"bufio"
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

// IDMapper gestiona el mapeo thread-safe de IDs a índices secuenciales
type IDMapper struct {
	mu      sync.RWMutex
	mapping map[int]int
	nextIdx int
	changed bool
}

// NewIDMapper crea un nuevo mapeador con el mapa inicial cargado
func NewIDMapper(initialMap map[int]int) *IDMapper {
	maxIdx := -1
	for _, idx := range initialMap {
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	return &IDMapper{
		mapping: initialMap,
		nextIdx: maxIdx + 1,
		changed: false,
	}
}

// Get obtiene el índice mapeado, o -1 si no existe
func (m *IDMapper) Get(id int) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if idx, ok := m.mapping[id]; ok {
		return idx
	}
	return -1
}

// GetOrCreate obtiene el índice existente o crea uno nuevo
func (m *IDMapper) GetOrCreate(id int) int {
	// Primero intentar lectura sin bloqueo de escritura
	m.mu.RLock()
	if idx, ok := m.mapping[id]; ok {
		m.mu.RUnlock()
		return idx
	}
	m.mu.RUnlock()

	// Si no existe, adquirir lock de escritura
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check en caso de que otro goroutine ya lo creó
	if idx, ok := m.mapping[id]; ok {
		return idx
	}

	// Crear nuevo mapeo
	idx := m.nextIdx
	m.mapping[id] = idx
	m.nextIdx++
	m.changed = true
	return idx
}

// HasChanged indica si el mapa fue modificado
func (m *IDMapper) HasChanged() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.changed
}

// GetMapping devuelve una copia del mapa actual
func (m *IDMapper) GetMapping() map[int]int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copy := make(map[int]int, len(m.mapping))
	for k, v := range m.mapping {
		copy[k] = v
	}
	return copy
}

// Count devuelve el número de mapeos
func (m *IDMapper) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.mapping)
}

// SaveItemMap guarda el mapeo movieId -> iIdx a un archivo CSV
func SaveItemMap(path string, itemMap map[int]int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(bufio.NewWriter(f))
	defer w.Flush()

	// Escribir header
	if err := w.Write([]string{"movieId", "iIdx"}); err != nil {
		return err
	}

	// Ordenar por iIdx para mantener consistencia
	type kv struct {
		movieId int
		iIdx    int
	}
	items := make([]kv, 0, len(itemMap))
	for movieId, iIdx := range itemMap {
		items = append(items, kv{movieId, iIdx})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].iIdx < items[j].iIdx
	})

	// Escribir filas
	for _, item := range items {
		if err := w.Write([]string{
			strconv.Itoa(item.movieId),
			strconv.Itoa(item.iIdx),
		}); err != nil {
			return err
		}
	}

	return nil
}

// SaveUserMap guarda el mapeo userId -> uIdx a un archivo CSV
func SaveUserMap(path string, userMap map[int]int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(bufio.NewWriter(f))
	defer w.Flush()

	// Escribir header
	if err := w.Write([]string{"userId", "uIdx"}); err != nil {
		return err
	}

	// Ordenar por uIdx para mantener consistencia
	type kv struct {
		userId int
		uIdx   int
	}
	users := make([]kv, 0, len(userMap))
	for userId, uIdx := range userMap {
		users = append(users, kv{userId, uIdx})
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].uIdx < users[j].uIdx
	})

	// Escribir filas
	for _, user := range users {
		if err := w.Write([]string{
			strconv.Itoa(user.userId),
			strconv.Itoa(user.uIdx),
		}); err != nil {
			return err
		}
	}

	return nil
}
