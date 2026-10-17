package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	phonebook "github.com/BeJloHukku/phonebook/internal/models"
	"github.com/BeJloHukku/phonebook/internal/storage/postgres"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Не удалось отправить ответ: %v", err)
	}
}

type CreateContactRequest struct {
	Name       string   `json:"name"`
	Surname    string   `json:"surname"`
	SecondName string   `json:"second_name"`
	Numbers    []string `json:"numbers"`
}

type UpdateContactRequest struct {
	Name       *string   `json:"name"`
	Surname    *string   `json:"surname"`
	SecondName *string   `json:"second_name"`
	Numbers    *[]string `json:"numbers"`
}

type Handler struct {
	store *postgres.Store
}

func (h *Handler) createContact(w http.ResponseWriter, r *http.Request) {
	var req CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "Имя обязательно", http.StatusUnprocessableEntity)
		return
	}
	if len(req.Numbers) == 0 {
		http.Error(w, "Должен быть хотя бы один номер", http.StatusUnprocessableEntity)
		return
	}

	contact := phonebook.Contact{
		Name:       req.Name,
		Surname:    req.Surname,
		SecondName: req.SecondName,
		Numbers:    req.Numbers,
	}

	created, err := h.store.CreateContact(r.Context(), contact)
	if err != nil {
		log.Printf("Не удалось создать контакт: %v", err)
		http.Error(w, "Ошибка создания контакта", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(created); err != nil {
		log.Printf("Не удалось отправить ответ: %v", err)
	}
}

func (h *Handler) getContactByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Котнакт не найден", http.StatusNotFound)
		return
	}
	c, err := h.store.GetContactByID(r.Context(), intID)
	if err != nil {
		log.Printf("Не удалось получить контакт: %v", err)
		http.Error(w, "Ошибка", http.StatusInternalServerError)
		return
	}
	if c.ID == 0 {
		http.Error(w, "Контакт не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(c); err != nil {
		log.Printf("Не удалось отправить ответ: %v", err)
	}
}

func (h *Handler) deleteContactByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Неверный формат ввода", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteContactByID(r.Context(), intID); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			http.Error(w, "Контакт не найден", http.StatusNotFound)
			return
		}
		log.Printf("Не удалось удалить контакт: %v", err)
		http.Error(w, "Ошибка удаления контакте", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateContact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Неверный формат ID", http.StatusBadRequest)
		return
	}

	var req UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	params := postgres.UpdateContactParams{
		Name:       req.Name,
		Surname:    req.Surname,
		SecondName: req.SecondName,
		Numbers:    req.Numbers,
	}

	if err := h.store.UpdateContact(r.Context(), intID, params); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			http.Error(w, "Контакт не найден", http.StatusNotFound)
			return
		}
		log.Printf("Не удалось обновить контакт: %v", err)
		http.Error(w, "Ошибка обновления контакта", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
