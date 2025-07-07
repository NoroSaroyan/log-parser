package v1

import (
	"log"
	"log-parser/internal/services/downloadinfo"
	"net/http"
)

// DownloadHandler оборачивает сервис
type DownloadHandler struct {
	svc downloadinfo.DownloadInfoService
}

// NewDownloadHandler конструктор
func NewDownloadHandler(svc downloadinfo.DownloadInfoService) *DownloadHandler {
	return &DownloadHandler{svc: svc}
}

// Get обрабатывает GET /download?pcbanumber=
func (h *DownloadHandler) Get(w http.ResponseWriter, r *http.Request) {
	pcba := r.URL.Query().Get("pcbanumber")
	if pcba == "" {
		respondError(w, http.StatusBadRequest, "pcbanumber is required")
		return
	}

	dto, err := h.svc.GetByPCBANumber(r.Context(), pcba)
	if err != nil {
		log.Printf("GetByPCBANumber error: %v", err)
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if dto.TcuPCBANumber == "" {
		respondError(w, http.StatusNotFound, "not found")
		return
	}

	respondJSON(w, http.StatusOK, dto)
}
