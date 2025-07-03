package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/albyma98/WASAText/service/api/reqcontext"
	"github.com/albyma98/WASAText/service/database"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o invalido"}`, http.StatusUnauthorized)
		return
	}

	idStr := ps.ByName("id")
	messageID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID non valido"}`, http.StatusBadRequest)
		return
	}

	// Decodifica body
	var body struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Emoji == "" {
		http.Error(w, `{"error":"Emoji non valida"}`, http.StatusBadRequest)
		return
	}

	// Recupera messaggio
	msg, err := rt.db.GetMessageByID(messageID)
	if err != nil {
		http.Error(w, `{"error":"Messaggio non trovato"}`, http.StatusNotFound)
		return
	}

	// Controlla se l’utente è membro della conversazione
	isMember, err := rt.db.IsMember(ctx.UserUUID, msg.IDConversation)
	if err != nil {
		http.Error(w, `{"error":"Errore DB"}`, http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, `{"error":"Utente non autorizzato"}`, http.StatusForbidden)
		return
	}

	// Controlla se ha già reagito
	reazioni, err := rt.db.GetReactionsByMessageID(messageID)

	if err != nil {
		http.Error(w, `{"error":"Errore nel recupero reazioni"}`, http.StatusInternalServerError)
		return
	}
	for _, r := range reazioni {
		if r.UUIDUser == ctx.UserUUID {
			http.Error(w, `{"error":"Reazione già presente"}`, http.StatusBadRequest)
			return
		}
	}

	// Aggiungi la reazione
	err = rt.db.AddReaction(messageID, ctx.UserUUID, body.Emoji)
	if err != nil {
		http.Error(w, `{"error":"Errore durante l'aggiunta della reazione"}`, http.StatusInternalServerError)
		return
	}

	user, err := rt.db.GetUserByUUID(ctx.UserUUID)
	if err != nil {
		http.Error(w, `{"error":"Errore DB"}`, http.StatusInternalServerError)
		return
	}
	// Risposta 201
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(database.ReactionWithUser{
		UUIDUser: ctx.UserUUID,
		Username: user.Username,
		Emoji:    body.Emoji,
	}); err != nil {
		http.Error(w, "errore nella codifica della risposta", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Verifica autenticazione
	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o non valido"}`, http.StatusUnauthorized)
		return
	}

	// 2. Estrai ID del messaggio
	idStr := ps.ByName("id")
	messageID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID non valido"}`, http.StatusBadRequest)
		return
	}

	// 3. Recupera il messaggio per avere la conversazione
	msg, err := rt.db.GetMessageByID(messageID)
	if err != nil {
		http.Error(w, `{"error":"Messaggio non trovato"}`, http.StatusNotFound)
		return
	}

	// 4. Verifica se l’utente fa parte della conversazione
	isMember, err := rt.db.IsMember(ctx.UserUUID, msg.IDConversation)
	if err != nil {
		http.Error(w, `{"error":"Errore DB"}`, http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, `{"error":"Utente non autorizzato"}`, http.StatusForbidden)
		return
	}

	// 5. Verifica se l’utente aveva una reazione
	reactions, err := rt.db.GetReactionsByMessageID(messageID)
	if err != nil {
		http.Error(w, `{"error":"Errore DB"}`, http.StatusInternalServerError)
		return
	}

	found := false
	for _, r := range reactions {
		if r.UUIDUser == ctx.UserUUID {
			found = true
			break
		}
	}
	if !found {
		http.Error(w, `{"error":"Reazione non trovata"}`, http.StatusNotFound)
		return
	}

	// 6. Rimuovi la reazione
	err = rt.db.RemoveReaction(messageID, ctx.UserUUID)
	if err != nil {
		http.Error(w, `{"error":"Errore durante la rimozione"}`, http.StatusInternalServerError)
		return
	}

	// 7. Risposta 204
	w.WriteHeader(http.StatusNoContent)
}
