package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/albyma98/WASAText/service/api/reqcontext"
	"github.com/albyma98/WASAText/service/database"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// ID conversazione dal path
	convID, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID conversazione non valido"}`, http.StatusBadRequest)
		return
	}

	// Verifica che la conversazione esista
	_, err = rt.db.GetConversationByID(convID)
	if err != nil {
		http.Error(w, `{"error":"Conversazione non trovata"}`, http.StatusNotFound)
		return
	}

	// Verifica che l’utente sia membro della conversazione
	isMember, err := rt.db.IsMember(ctx.UserUUID, convID)
	if err != nil {
		http.Error(w, `{"error":"Errore interno durante il controllo membri"}`, http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, `{"error":"L’utente non fa parte della conversazione"}`, http.StatusForbidden)
		return
	}

	// Decode JSON request
	var body struct {
		Type        string  `json:"type"`
		Content     *string `json:"content"`
		MediaUrl    *string `json:"mediaUrl"`
		IDRepliesTo *int64  `json:"idRepliesTo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"JSON non valido"}`, http.StatusBadRequest)
		return
	}

	// Validazioni logiche
	if body.Type == "text" && (body.Content == nil || *body.Content == "") {
		http.Error(w, `{"error":"Content richiesto per messaggio testuale"}`, http.StatusBadRequest)
		return
	}
	if body.Type == "photo" && (body.MediaUrl == nil || *body.MediaUrl == "") {
		http.Error(w, `{"error":"MediaUrl richiesto per messaggio foto"}`, http.StatusBadRequest)
		return
	}
	if body.Type != "text" && body.Type != "photo" {
		http.Error(w, `{"error":"Tipo di messaggio non valido"}`, http.StatusBadRequest)
		return
	}

	// Se il messaggio è una reply, verifica che il messaggio esista
	if body.IDRepliesTo != nil {
		if _, err := rt.db.GetMessageByID(*body.IDRepliesTo); err != nil {
			http.Error(w, `{"error":"Messaggio a cui rispondere non trovato"}`, http.StatusNotFound)
			return
		}
	}

	// Crea struttura Message
	msg := database.Message{
		Type:           body.Type,
		Content:        "",
		MediaUrl:       body.MediaUrl,
		IDConversation: convID,
		UUIDSender:     ctx.UserUUID,
		IDRepliesTo:    body.IDRepliesTo,
	}
	if body.Content != nil {
		msg.Content = *body.Content
	}

	// Inserisci messaggio
	newID, err := rt.db.CreateMessage(msg)
	if err != nil {
		http.Error(w, `{"error":"Errore durante la creazione del messaggio"}`, http.StatusInternalServerError)
		return
	}

	// Recupera messaggio completo (incluso timestamp)
	msg.ID = newID
	msg.Timestamp = time.Now().Format(time.RFC3339)

	// Risposta
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, "errore nella codifica della risposta", http.StatusInternalServerError)
		return
	}

}

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Autenticazione
	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o non valido"}`, http.StatusUnauthorized)
		return
	}

	// 1. Estrai ID dal path
	idStr := ps.ByName("id")
	msgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	// 2. Estrai UUID utente autenticato dal context
	uuid := ctx.UserUUID

	// 3. Elimina messaggio (solo se inviato da lui)
	err = rt.db.DeleteMessageByID(msgID, uuid)
	if err != nil {
		if err.Error() == "nessun messaggio eliminato (ID non esistente o UUID non corrispondente)" {
			http.Error(w, `{"error":"Non autorizzato o messaggio non trovato"}`, http.StatusForbidden)
		} else {
			http.Error(w, `{"error":"Errore durante l'eliminazione"}`, http.StatusInternalServerError)
		}
		return
	}

	// 4. Risposta 204
	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Estrai l'ID del messaggio dalla path
	idMsg, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil || idMsg <= 0 {
		http.Error(w, `{"error":"ID messaggio non valido"}`, http.StatusBadRequest)
		return
	}

	// 2. Decodifica il body JSON
	var body struct {
		IdConversation int64 `json:"idConversation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Body malformato"}`, http.StatusBadRequest)
		return
	}
	if body.IdConversation <= 0 {
		http.Error(w, `{"error":"ID conversazione mancante o non valido"}`, http.StatusBadRequest)
		return
	}

	// 3. Esegui l'inoltro del messaggio
	newID, err := rt.db.ForwardMessage(idMsg, body.IdConversation, ctx.UserUUID)
	if err != nil {
		switch err.Error() {
		case "messaggio originale non trovato":
			http.Error(w, `{"error":"Messaggio non trovato"}`, http.StatusNotFound)
		case "utente non autorizzato":
			http.Error(w, `{"error":"Non puoi inoltrare in questa conversazione"}`, http.StatusForbidden)
		default:
			http.Error(w, `{"error":"Errore interno"}`, http.StatusInternalServerError)
		}
		return
	}

	// 4. Recupera il messaggio appena creato per inviarlo come risposta
	forwardedMsg, err := rt.db.GetMessageByID(newID)
	if err != nil {
		http.Error(w, `{"error":"Errore nel recupero del messaggio inoltrato"}`, http.StatusInternalServerError)
		return
	}

	// 5. Risposta 201 con JSON del messaggio
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(forwardedMsg); err != nil {
		http.Error(w, "errore nella codifica della risposta", http.StatusInternalServerError)
		return
	}

}
