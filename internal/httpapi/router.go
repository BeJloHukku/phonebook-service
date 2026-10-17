package httpapi

import (
	"net/http"

	"github.com/BeJloHukku/phonebook/internal/storage/postgres"
)

func NewRouter(store *postgres.Store) http.Handler {
	h := &Handler{store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /contact", h.createContact)
	mux.HandleFunc("GET /contact/{id}", h.getContactByID)
	mux.HandleFunc("DELETE /contact/{id}", h.deleteContactByID)
	mux.HandleFunc("PATCH /contact/{id}", h.updateContact)

	return mux
}
