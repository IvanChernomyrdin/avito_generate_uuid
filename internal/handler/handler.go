package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

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

func (h *Handler) GenerateKeysHandler(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) ValidateKeyHandler(w http.ResponseWriter, r *http.Request) {}
