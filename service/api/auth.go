package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/albyma98/WASAText/service/database"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

// Handler per POST /session (login o registrazione)
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Username string `json:"username"`
	}

	// Decodifica JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Malformed request"}`, http.StatusBadRequest)
		return
	}

	// Validazione sintattica secondo OpenAPI
	if len(req.Username) < 3 || len(req.Username) > 16 {
		http.Error(w, `{"error":"Username must be between 3 and 16 characters"}`, http.StatusBadRequest)
		return
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(req.Username) {
		http.Error(w, `{"error":"Username can only contain letters, digits and underscores"}`, http.StatusBadRequest)
		return
	}

	// Cerca utente esistente
	users, err := rt.db.SearchUsersByPrefix(req.Username)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	for _, u := range users {
		if u.Username == req.Username {
			w.WriteHeader(http.StatusOK) // 200
			if err := json.NewEncoder(w).Encode(u); err != nil {
				http.Error(w, "errore nella codifica della risposta", http.StatusInternalServerError)
				return
			}
			return
		}
	}

	// Utente non esiste → creazione
	newUUID, err := uuid.NewV4()
	if err != nil {
		http.Error(w, `{"error":"Failed to generate UUID"}`, http.StatusInternalServerError)
		return
	}
	err = rt.db.CreateUser(newUUID.String(), req.Username, "")
	if err != nil {
		http.Error(w, `{"error":"Unable to create user"}`, http.StatusInternalServerError)
		return
	}

	user := database.User{
		UUID:     newUUID.String(),
		Username: req.Username,
		PhotoUrl: nil,
	}

	w.WriteHeader(http.StatusCreated) // 201
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "errore nella codifica della risposta", http.StatusInternalServerError)
		return
	}
}
