package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/albyma98/WASAText/service/api/reqcontext"
	"github.com/albyma98/WASAText/service/database"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	convs, err := rt.db.GetConversationsByUser(ctx.UserUUID)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	type ResponseConversation struct {
		ID                   int64   `json:"id"`
		IsDirect             bool    `json:"isDirect"`
		GroupName            *string `json:"groupName"`
		GroupPhoto           *string `json:"groupPhoto"`
		TimestampCreated     string  `json:"timestampCreated"`
		TimestampLastMessage string  `json:"timestampLastMessage"`

		PeerUsername    *string `json:"peerUsername,omitempty"`
		PeerPhoto       *string `json:"peerPhoto,omitempty"`
		LastMessageText *string `json:"lastMessageText,omitempty"`
		LastMessageType *string `json:"lastMessageType,omitempty"`
	}

	var output []ResponseConversation

	for _, c := range convs {
		item := ResponseConversation{
			ID:                   c.ID,
			IsDirect:             c.IsDirect,
			GroupName:            c.GroupName,
			GroupPhoto:           c.GroupPhoto,
			TimestampCreated:     c.TimestampCreated,
			TimestampLastMessage: c.TimestampLastMessage,
		}

		// 1. Ottieni l'ultimo messaggio (se esiste)
		if lastMsg, err := rt.db.GetLastMessage(c.ID); err == nil {
			item.LastMessageText = &lastMsg.Content
			item.LastMessageType = &lastMsg.Type
		}

		// 2. Se è diretta, ottieni info dell'altro utente
		if c.IsDirect {
			if peer, err := rt.db.GetPeerData(c.ID, ctx.UserUUID); err == nil {
				item.PeerUsername = &peer.Username
				item.PeerPhoto = peer.PhotoUrl
			}
		}

		output = append(output, item)
	}

	errs := json.NewEncoder(w).Encode(map[string]interface{}{
		"conversations": output,
	})
	if errs != nil {
		log.Printf("Errore nella codifica JSON: %v", errs)
		http.Error(w, "Errore interno del server", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o non valido"}`, http.StatusUnauthorized)
		return
	}

	var body struct {
		IsDirect   bool     `json:"isDirect"`
		GroupName  *string  `json:"groupName"`
		GroupPhoto *string  `json:"groupPhoto"`
		Members    []string `json:"members"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Corpo della richiesta non valido"}`, http.StatusBadRequest)
		return
	}

	if body.IsDirect {
		// Validazione: deve esserci un solo membro e groupName/groupPhoto devono essere null
		if len(body.Members) != 1 || body.GroupName != nil || body.GroupPhoto != nil {
			http.Error(w, `{"error":"Parametri non validi per conversazione diretta"}`, http.StatusBadRequest)
			return
		}

		// Controlla se già esiste
		_, err := rt.db.GetDirectConversationBetween(ctx.UserUUID, body.Members[0])
		if err == nil {
			// Se non dà errore è perché la conversazione esiste già
			http.Error(w, `{"error":"Conversazione già esistente"}`, http.StatusConflict)
			return
		}

		conv, err := rt.db.CreateDirectConversation(ctx.UserUUID, body.Members[0])
		if err != nil {
			http.Error(w, `{"error":"Errore nella creazione della conversazione"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(conv); err != nil {
			http.Error(w, `{"error":" errore nella codifica di risposta"}`, http.StatusInternalServerError)
			return
		}
		return

	} else {
		// Validazione parametri gruppo
		if body.GroupName == nil || len(*body.GroupName) < 3 || len(body.Members) < 1 {
			http.Error(w, `{"error":"Dati gruppo non validi"}`, http.StatusBadRequest)
			return
		}

		// Crea conversazione di gruppo (solo il creatore viene aggiunto, gli altri saranno aggiunti separatamente)
		conv, err := rt.db.CreateGroupConversation(ctx.UserUUID, body.GroupName, body.GroupPhoto)
		if err != nil {
			http.Error(w, `{"error":"Errore nella creazione del gruppo"}`, http.StatusInternalServerError)
			return
		}

		// Inserisci gli altri membri specificati
		for _, uuid := range body.Members {
			if uuid == ctx.UserUUID {
				// Il creatore è già stato inserito
				continue
			}

			if err := rt.db.AddMember(uuid, conv.ID); err != nil {
				http.Error(w, `{"error":"Errore aggiunta membro"}`, http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(conv); err != nil {
			http.Error(w, `{"error":" errore nella codifica di risposta"}`, http.StatusInternalServerError)
			return
		}
		return
	}
}

func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o non valido"}`, http.StatusUnauthorized)
		return
	}

	// Prendi l’ID della conversazione da path param
	convID, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID conversazione non valido"}`, http.StatusBadRequest)
		return
	}

	// Recupera la conversazione
	conv, err := rt.db.GetConversationByID(convID)
	if err != nil {
		http.Error(w, `{"error":"Conversazione non trovata"}`, http.StatusNotFound)
		return
	}

	// Controlla se l’utente ne fa parte
	isMember, err := rt.db.IsMember(ctx.UserUUID, convID)
	if err != nil {
		http.Error(w, `{"error":"Errore accesso conversazione"}`, http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, `{"error":"Accesso negato alla conversazione"}`, http.StatusForbidden)
		return
	}

	// Recupera i messaggi della conversazione comprensivi delle reazioni
	baseMessages, err := rt.db.GetMessagesByConversationID(convID)
	if err != nil {
		http.Error(w, `{"error":"Errore recupero messaggi"}`, http.StatusInternalServerError)
		return
	}

	// Aggiungi gli status delivered/seen ad ogni messaggio

	type ReplyMessage struct {
		Type     string  `json:"type"`
		Content  string  `json:"content"`
		MediaUrl *string `json:"mediaUrl"`
	}


	type MessageWithStatus struct {
		database.Message
		Delivered      []string      `json:"delivered"`
		Seen           []string      `json:"seen"`
		UsernameSender string        `json:"usernameSender"`
		ReplyToMessage *ReplyMessage `json:"replyToMessage,omitempty"`
	}

	var messagesWithStatus []MessageWithStatus
	for _, m := range baseMessages {
		statuses, err := rt.db.GetAllStatusesByMessage(m.ID)
		if err != nil {
			http.Error(w, `{"error":"Errore recupero stati messaggi"}`, http.StatusInternalServerError)
			return
		}

		var delivered []string
		var seen []string
		for _, st := range statuses {
			if st.Delivered {
				delivered = append(delivered, st.UUIDUser)
			}
			if st.Seen {
				seen = append(seen, st.UUIDUser)
			}
		}

		username := ""
		if user, err := rt.db.GetUserByUUID(m.UUIDSender); err == nil {
			username = user.Username
		}


		var replyMsg *ReplyMessage
		if m.IDRepliesTo != nil {
			if original, err := rt.db.GetMessageByID(*m.IDRepliesTo); err == nil {
				replyMsg = &ReplyMessage{
					Type:     original.Type,
					Content:  original.Content,
					MediaUrl: original.MediaUrl,
				}
			}
		}

		messagesWithStatus = append(messagesWithStatus, MessageWithStatus{
			Message:        m,
			Delivered:      delivered,
			Seen:           seen,
			UsernameSender: username,
			ReplyToMessage: replyMsg,
		})
	}

	// Se diretta, recupera info del peer
	var usernamePeer *string
	var photoUrlPeer *string
	if conv.IsDirect {
		peer, err := rt.db.GetPeerData(convID, ctx.UserUUID)
		if err != nil {
			http.Error(w, `{"error":"Errore recupero peer"}`, http.StatusInternalServerError)
			return
		}
		usernamePeer = &peer.Username
		photoUrlPeer = peer.PhotoUrl
	}

	// Numero di membri della conversazione
	members, err := rt.db.GetMembersByConversation(convID)
	if err != nil {
		http.Error(w, `{"error":"Errore recupero membri"}`, http.StatusInternalServerError)
		return
	}

	// Dettagli conversazione da restituire
	type conversationDetail struct {
		ID            int64   `json:"id"`
		IsDirect      bool    `json:"isDirect"`
		GroupName     *string `json:"groupName"`
		GroupPhoto    *string `json:"groupPhoto"`
		UsernamePeer  *string `json:"usernamePeer,omitempty"`
		PhotoUrlPeer  *string `json:"photoUrlPeer,omitempty"`
		NumberMembers int     `json:"numberMembers"`
	}

	convDetail := conversationDetail{
		ID:            conv.ID,
		IsDirect:      conv.IsDirect,
		GroupName:     conv.GroupName,
		GroupPhoto:    conv.GroupPhoto,
		UsernamePeer:  usernamePeer,
		PhotoUrlPeer:  photoUrlPeer,
		NumberMembers: len(members),
	}

	// Tutto ok, restituisci dettagli e messaggi
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"conversationDetail": convDetail,
		"messages":           messagesWithStatus,
	}); err != nil {
		http.Error(w, `{"error":" errore nella codifica di risposta"}`, http.StatusInternalServerError)
		return
	}
}

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o non valido"}`, http.StatusUnauthorized)
		return
	}

	// Prendi ID da path
	convID, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID conversazione non valido"}`, http.StatusBadRequest)
		return
	}

	// Controlla se la conversazione esiste
	conv, err := rt.db.GetConversationByID(convID)
	if err != nil {
		http.Error(w, `{"error":"Conversazione non trovata"}`, http.StatusNotFound)
		return
	}

	// Solo per gruppi
	if conv.IsDirect {
		http.Error(w, `{"error":"Non puoi modificare una conversazione diretta"}`, http.StatusBadRequest)
		return
	}

	// Controlla che l’utente sia membro
	isMember, err := rt.db.IsMember(ctx.UserUUID, convID)
	if err != nil {
		http.Error(w, `{"error":"Errore accesso"}`, http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, `{"error":"Non hai accesso alla conversazione"}`, http.StatusForbidden)
		return
	}

	// Leggi i dati del body
	var body struct {
		GroupName string `json:"groupName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Body malformato"}`, http.StatusBadRequest)
		return
	}

	// Controllo obbligatorietà e validità
	if body.GroupName == "" {
		http.Error(w, `{"error":"Nome gruppo obbligatorio"}`, http.StatusBadRequest)
		return
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
	if !re.MatchString(body.GroupName) {
		http.Error(w, `{"error":"Nome gruppo non valido"}`, http.StatusBadRequest)
		return
	}

	// Esegui update nel DB
	err = rt.db.SetGroupName(convID, body.GroupName)
	if err != nil {
		http.Error(w, `{"error":"Errore aggiornamento gruppo"}`, http.StatusInternalServerError)
		return
	}

	// Ritorna l’oggetto aggiornato
	updated, err := rt.db.GetConversationByID(convID)
	if err != nil {
		http.Error(w, `{"error":"Errore lettura dopo aggiornamento"}`, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(updated); err != nil {
		http.Error(w, `{"error":" errore nella codifica di risposta"}`, http.StatusInternalServerError)
		return
	}

}

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o non valido"}`, http.StatusUnauthorized)
		return
	}

	// Prendi ID da path
	convID, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID conversazione non valido"}`, http.StatusBadRequest)
		return
	}

	// Controlla se la conversazione esiste
	conv, err := rt.db.GetConversationByID(convID)
	if err != nil {
		http.Error(w, `{"error":"Conversazione non trovata"}`, http.StatusNotFound)
		return
	}

	// Solo per gruppi
	if conv.IsDirect {
		http.Error(w, `{"error":"Non puoi modificare una conversazione diretta"}`, http.StatusBadRequest)
		return
	}

	// Controlla che l’utente sia membro
	isMember, err := rt.db.IsMember(ctx.UserUUID, convID)
	if err != nil {
		http.Error(w, `{"error":"Errore accesso"}`, http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, `{"error":"Non hai accesso alla conversazione"}`, http.StatusForbidden)
		return
	}

	// Leggi i dati del body
	// Crea ./webui/public/ se non esiste
	if err := os.MkdirAll("./webui/public", os.ModePerm); err != nil {
		log.Println("❌ Errore creazione cartella ./webui/public:", err)
		http.Error(w, `{"error":"Cannot create upload directory"}`, http.StatusInternalServerError)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Println("❌ Errore ParseMultipartForm:", err)
		http.Error(w, `{"error":"Cannot parse form data"}`, http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("photo")
	if err != nil {
		log.Println("⚠️ ERRORE R.FormFile:", err)
		http.Error(w, `{"error":"File not found in request"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	safeName := regexp.MustCompile(`[^a-zA-Z0-9\.\-_]`).ReplaceAllString(handler.Filename, "_")
	filename := fmt.Sprintf("conv%d_%d_%s", convID, time.Now().Unix(), safeName)
	filepath := "./webui/public/" + filename

	dst, err := os.Create(filepath)
	if err != nil {
		log.Println("❌ Errore salvataggio file:", err)
		http.Error(w, `{"error":"Cannot save file"}`, http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Println("❌ Errore copia file:", err)
		http.Error(w, `{"error":"Error saving file"}`, http.StatusInternalServerError)
		return
	}

	publicPath := "/" + filename

	err = rt.db.SetGroupPhoto(convID, publicPath)
	if err != nil {
		http.Error(w, `{"error":"Errore aggiornamento gruppo"}`, http.StatusInternalServerError)
		return
	}

	// Ritorna l’oggetto aggiornato
	updated, err := rt.db.GetConversationByID(convID)
	if err != nil {
		http.Error(w, `{"error":"Errore lettura dopo aggiornamento"}`, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(updated); err != nil {
		http.Error(w, `{"error":" errore nella codifica di risposta"}`, http.StatusInternalServerError)
		return
	}

}

func (rt *_router) getGroupMembers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o non valido"}`, http.StatusUnauthorized)
		return
	}

	convID, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID conversazione non valido"}`, http.StatusBadRequest)
		return
	}

	if _, err := rt.db.GetConversationByID(convID); err != nil {
		http.Error(w, `{"error":"Conversazione non trovata"}`, http.StatusNotFound)
		return
	}

	isMember, err := rt.db.IsMember(ctx.UserUUID, convID)
	if err != nil {
		http.Error(w, `{"error":"Errore accesso DB"}`, http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, `{"error":"Non fai parte della conversazione"}`, http.StatusForbidden)
		return
	}

	uuids, err := rt.db.GetMembersByConversation(convID)
	if err != nil {
		http.Error(w, `{"error":"Errore recupero membri"}`, http.StatusInternalServerError)
		return
	}

	var usernames []string
	for _, uuid := range uuids {
		user, err := rt.db.GetUserByUUID(uuid)
		if err != nil {
			http.Error(w, `{"error":"Errore recupero utente"}`, http.StatusInternalServerError)
			return
		}
		usernames = append(usernames, user.Username)
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"members": usernames,
	}); err != nil {
		http.Error(w, `{"error":"Errore codifica risposta"}`, http.StatusInternalServerError)
		return
	}
}

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Controlla autenticazione
	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o invalido"}`, http.StatusUnauthorized)
		return
	}

	// Prendi ID conversazione dal path
	convID, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID conversazione non valido"}`, http.StatusBadRequest)
		return
	}

	// Verifica che la conversazione esista e sia di tipo gruppo
	conversation, err := rt.db.GetConversationByID(convID)
	if err != nil {
		http.Error(w, `{"error":"Conversazione non trovata"}`, http.StatusNotFound)
		return
	}
	if conversation.IsDirect {
		http.Error(w, `{"error":"Non puoi aggiungere membri a una conversazione diretta"}`, http.StatusBadRequest)
		return
	}

	// Verifica che l’utente sia membro del gruppo
	isMember, err := rt.db.IsMember(ctx.UserUUID, convID)
	if err != nil {
		http.Error(w, `{"error":"Errore verifica membro"}`, http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, `{"error":"Non fai parte della conversazione"}`, http.StatusForbidden)
		return
	}

	// Leggi il corpo della richiesta
	var body struct {
		Members []string `json:"members"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Members) == 0 {
		http.Error(w, `{"error":"Payload non valido o membri mancanti"}`, http.StatusBadRequest)
		return
	}

	var added []string
	var alreadyPresent []string

	for _, uuid := range body.Members {
		isAlready, err := rt.db.IsMember(uuid, convID)
		if err != nil {
			http.Error(w, `{"error":"Errore accesso al DB"}`, http.StatusInternalServerError)
			return
		}
		if isAlready {
			alreadyPresent = append(alreadyPresent, uuid)
			continue
		}

		// Aggiungi il membro
		err = rt.db.AddMember(uuid, convID)
		if err != nil {
			http.Error(w, `{"error":"Errore aggiunta membro"}`, http.StatusInternalServerError)
			return
		}
		added = append(added, uuid)
	}

	// Risposta finale	convs); err != nil {

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"added":          added,
		"alreadyPresent": alreadyPresent,
	}); err != nil {
		http.Error(w, `{"error":" errore nella codifica di risposta"}`, http.StatusInternalServerError)
		return
	}
}

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Token mancante o invalido"}`, http.StatusUnauthorized)
		return
	}

	// Parse ID conversazione
	conversationID, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"ID conversazione non valido"}`, http.StatusBadRequest)
		return
	}

	// Verifica esistenza e tipo della conversazione
	conversation, err := rt.db.GetConversationByID(conversationID)
	if err != nil {
		http.Error(w, `{"error":"Conversazione non trovata"}`, http.StatusNotFound)
		return
	}
	if conversation.IsDirect {
		http.Error(w, `{"error":"Una conversazione diretta non può essere abbandonata"}`, http.StatusBadRequest)
		return
	}

	// Verifica che l’utente sia membro
	isMember, err := rt.db.IsMember(ctx.UserUUID, conversationID)
	if err != nil {
		http.Error(w, `{"error":"Errore controllo membro"}`, http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, `{"error":"Non sei membro di questa conversazione"}`, http.StatusForbidden)
		return
	}

	// Rimuovi il membro
	err = rt.db.RemoveMember(ctx.UserUUID, conversationID)
	if err != nil {
		http.Error(w, `{"error":"Errore rimozione membro"}`, http.StatusInternalServerError)
		return
	}

	// Se la conversazione è vuota, eliminala
	err = rt.db.DeleteConversationIfEmpty(conversationID)
	if err != nil {
		http.Error(w, `{"error":"Errore eliminazione conversazione"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
