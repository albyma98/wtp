package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/albyma98/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) updateMessageStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o invalido"}`, http.StatusUnauthorized)
		return
	}

	// Estrai ID messaggio
	idStr := ps.ByName("id")
	messageID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID non valido"}`, http.StatusBadRequest)
		return
	}

	// Decodifica il body
	var req struct {
		Delivered bool `json:"delivered"`
		Seen      bool `json:"seen"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Formato JSON non valido"}`, http.StatusBadRequest)
		return
	}

	// Non si può mettere false
	if !req.Delivered && !req.Seen {
		http.Error(w, `{"error":"Non è possibile impostare delivered o seen a false"}`, http.StatusBadRequest)
		return
	}

	// Controlla se esiste lo status
	_, err = rt.db.GetMessageStatus(ctx.UserUUID, messageID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, `{"error":"Messaggio non trovato o stato non tracciabile"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Errore interno"}`, http.StatusInternalServerError)
		}
		return
	}

	// Aggiorna i campi richiesti
	if req.Seen {
		err = rt.db.SetSeen(ctx.UserUUID, messageID)
	} else if req.Delivered {
		err = rt.db.SetDelivered(ctx.UserUUID, messageID)
	}
	if err != nil {
		http.Error(w, `{"error":"Errore durante aggiornamento"}`, http.StatusInternalServerError)
		return
	}

	// Recupera lo status aggiornato
	status, err := rt.db.GetMessageStatus(ctx.UserUUID, messageID)
	if err != nil {
		http.Error(w, `{"error":"Errore nel recupero dello stato aggiornato"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "errore nella codifica della risposta", http.StatusInternalServerError)
		return
	}

}
