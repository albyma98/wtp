package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/albyma98/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// Handler per GET /user/me
func (rt *_router) getMyUserInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Verifica presenza del token (già validato da wrap)
	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Recupera utente autenticato
	user, err := rt.db.GetUserByUUID(ctx.UserUUID)
	if err != nil {
		http.Error(w, `{"error":"User not found or database error"}`, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "errore nella codifica della risposta", http.StatusInternalServerError)
		return
	}

}

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Username *string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Aggiorna username se presente
	if req.Username != nil {
		if len(*req.Username) < 3 || len(*req.Username) > 16 {
			http.Error(w, `{"error":"Username must be between 3 and 16 characters"}`, http.StatusBadRequest)
			return
		}
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(*req.Username) {
			http.Error(w, `{"error":"Username format invalid"}`, http.StatusBadRequest)
			return
		}

		// Verifica che non sia già in uso da un altro utente
		users, err := rt.db.SearchUsersByPrefix(*req.Username)
		if err != nil {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
			return
		}
		for _, u := range users {
			if u.Username == *req.Username && u.UUID != ctx.UserUUID {
				http.Error(w, `{"error":"Username already in use"}`, http.StatusConflict)
				return
			}
		}

		if err := rt.db.SetUserName(ctx.UserUUID, *req.Username); err != nil {
			http.Error(w, `{"error":"Unable to update username"}`, http.StatusInternalServerError)
			return
		}
	}

	// Recupera e restituisci utente aggiornato
	user, err := rt.db.GetUserByUUID(ctx.UserUUID)
	if err != nil {
		http.Error(w, `{"error":"Unable to retrieve updated user"}`, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "errore nella codifica della risposta", http.StatusInternalServerError)
		return
	}

}

func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	// Crea ./webui/public/ se non esiste (evita errore "no such file or directory")
	if err := os.MkdirAll("./webui/public", os.ModePerm); err != nil {
		log.Println("❌ Errore creazione cartella ./webui/public:", err)
		http.Error(w, `{"error":"Cannot create upload directory"}`, http.StatusInternalServerError)
		return
	}

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 🔍 DEBUG HEADERS E CONTENT-TYPE
	log.Println("🛂 Headers:", r.Header)
	log.Println("📦 Content-Type:", r.Header.Get("Content-Type"))

	err := r.ParseMultipartForm(10 << 20) // max 10MB
	if err != nil {
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

	// 🔍 LOG SUCCESSSO FILE
	log.Println("✅ FILE TROVATO:", handler.Filename, handler.Header.Get("Content-Type"))

	// Genera nome sicuro per il file
	safeName := regexp.MustCompile(`[^a-zA-Z0-9\.\-_]`).ReplaceAllString(handler.Filename, "_")
	filename := fmt.Sprintf("%s_%d_%s", ctx.UserUUID, time.Now().Unix(), safeName)
	filepath := "./webui/public/" + filename

	// Salva il file
	dst, err := os.Create(filepath)
	if err != nil {
		log.Println("❌ Errore salvataggio file:", err)
		http.Error(w, `{"error":"Cannot save file"}`, http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		log.Println("❌ Errore copia file:", err)
		http.Error(w, `{"error":"Error saving file"}`, http.StatusInternalServerError)
		return
	}

	// Path da salvare nel DB
	publicPath := "/" + filename

	// Salva nel DB
	if err := rt.db.SetPhotoUrl(ctx.UserUUID, publicPath); err != nil {
		log.Println("❌ Errore salvataggio DB:", err)
		http.Error(w, `{"error":"Unable to update photo URL"}`, http.StatusInternalServerError)
		return
	}

	// Restituisci utente aggiornato
	user, err := rt.db.GetUserByUUID(ctx.UserUUID)
	if err != nil {
		log.Println("❌ Errore recupero utente:", err)
		http.Error(w, `{"error":"Unable to retrieve updated user"}`, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Errore nell'encoding JSON della risposta user: %v", err)
		http.Error(w, "Errore nel generare la risposta", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) getAllUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	users, err := rt.db.GetAllUsers()
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "errore nella codifica della risposta", http.StatusInternalServerError)
		return
	}

}

func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	if ctx.UserUUID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Legge il parametro "search" dalla query string
	query := r.URL.Query().Get("search")
	if query == "" || len(query) < 1 || len(query) > 16 || !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(query) {
		http.Error(w, `{"error":"Invalid search parameter"}`, http.StatusBadRequest)
		return
	}

	users, err := rt.db.SearchUsersByPrefix(query)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "errore nella codifica della risposta", http.StatusInternalServerError)
		return
	}

}
