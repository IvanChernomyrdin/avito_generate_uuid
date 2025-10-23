package handler

import (
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
	"unicode"

	"github.com/IvanChernomyrdin/avito-key-generate/config/db"
	"github.com/IvanChernomyrdin/avito-key-generate/runtime/logger"
)

type Handler struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewHandler(db *sql.DB, logger *logger.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
	}
}

func (h *Handler) PingDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := db.PingDatabase(); err != nil {
		h.logger.Error("handler: PingDatabase", "Database ping failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database unavailable"})
		return
	}
	h.logger.Info("handler: PingDatabase", "Database ping successful")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK", "database": "connected"})
}

func (h *Handler) GetGroupsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(patternTemplates)
}

func (h *Handler) ValidateKeyHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Group string   `json:"group"`
		Keys  []string `json:"keys"`
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("handler", "Invalid JSON", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON format"})
		return
	}
	if request.Group == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Group is required"})
		return
	}
	if len(request.Keys) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Keys array is empty"})
		return
	}
	pattern, exists := patternTemplates[request.Group]
	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unknown group"})
	}

	validKeys := []string{}
	invalidKeys := []string{}

	for _, key := range request.Keys {
		if isValidKey(key, pattern) {
			validKeys = append(validKeys, key)
		} else {
			invalidKeys = append(invalidKeys, key)
		}
	}

	type ValidationResponse struct {
		Group        string   `json:"group"`
		Pattern      string   `json:"pattern"`
		TotalCount   int      `json:"total_count"`
		ValidCount   int      `json:"valid_count"`
		ValidKeys    []string `json:"valid_keys"`
		InvalidCount int      `json:"invalid_count"`
		InvalidKeys  []string `json:"invalid_keys"`
	}

	response := &ValidationResponse{
		Group:        request.Group,
		Pattern:      pattern,
		TotalCount:   len(request.Keys),
		ValidCount:   len(validKeys),
		ValidKeys:    validKeys,
		InvalidCount: len(invalidKeys),
		InvalidKeys:  invalidKeys,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func isValidKey(key, pattern string) bool {
	if len(key) != len(pattern) {
		return false
	}

	for i, char := range key {
		patternChar := pattern[i]

		switch patternChar {
		case 'X':
			if !isAlphaNumeric(byte(char)) {
				return false
			}
		default:
			if byte(char) != patternChar {
				return false
			}
		}
	}
	return true
}

func isAlphaNumeric(char byte) bool {
	return unicode.IsLetter(rune(char)) || unicode.IsNumber(rune(char))
}

func (h *Handler) GenerateKeysHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Group string
		Count int
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON format"})
		return
	}
	if request.Group == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Group is required"})
		return
	}
	if request.Count == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Count can't be empty or equal to 0"})
		return
	}

	pattern, exists := patternTemplates[request.Group]
	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unknown group"})
		return
	}

	// сгенерированные ключи и хэштаблица их для проверки что уже такой был
	generateKeys := make([]string, 0, request.Count)
	usedKeys := make(map[string]bool)

	for len(generateKeys) < request.Count {
		key := generateKey(pattern)

		if !usedKeys[key] {
			exists, err := h.checkKeyExists(key)
			if err != nil {
				h.logger.Error("handler", "Database error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			}
			if !exists {
				if err := h.saveKeyToDB(key, request.Group, pattern); err != nil {
					h.logger.Error("handler", "Failed to save key", err)
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save key"})
				}

				generateKeys = append(generateKeys, key)
				usedKeys[key] = true
			}
		}
	}

	type GenerateKey struct {
		Group          string   `json:"group"`
		Pattern        string   `json:"pattern"`
		Generate_count int      `json:"count"`
		Keys           []string `json:"keys"`
	}
	response := &GenerateKey{
		Group:          request.Group,
		Pattern:        pattern,
		Generate_count: len(generateKeys),
		Keys:           generateKeys,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func generateKey(pattern string) string {
	key := make([]byte, len(pattern))

	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case 'X':
			key[i] = randomAlphaNumeric()
		default:
			key[i] = pattern[i]
		}
	}
	return string(key)
}

func randomAlphaNumeric() byte {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return chars[rand.Intn(len(chars))]
}

func (h *Handler) checkKeyExists(key string) (bool, error) {
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM keys WHERE key_value = $1 AND status = TRUE)", key).Scan(&exists)
	return exists, err
}

func (h *Handler) saveKeyToDB(key, group, pattern string) error {
	_, err := h.db.Exec("INSERT INTO keys (key_value, group_name, pattern, status, created_at) VALUES ($1, $2, $3, $4, $5)", key, group, pattern, "TRUE", time.Now())
	return err
}
