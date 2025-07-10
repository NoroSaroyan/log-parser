package v1

import (
	"github.com/NoroSaroyan/log-parser/internal/services/downloadinfo"
	"log"
	"net/http"
)

// DownloadHandler provides HTTP handlers for working with DownloadInfo resources.
//
// It exposes endpoints for retrieving download information by PCBA number.
// The handler depends on a DownloadInfoService for business logic and data access.
type DownloadHandler struct {
	svc downloadinfo.DownloadInfoService
}

// NewDownloadHandler creates a new DownloadHandler with the provided DownloadInfoService.
//
// This constructor is typically used to wire up the handler in your router:
//
//	svc := downloadinfo.NewDownloadInfoService(...)
//	handler := v1.NewDownloadHandler(svc)
//	r.Get("/download", handler.Get)
func NewDownloadHandler(svc downloadinfo.DownloadInfoService) *DownloadHandler {
	return &DownloadHandler{svc: svc}
}

// Get handles HTTP GET requests for retrieving DownloadInfo by PCBA number.
//
// It reads the "pcbanumber" query parameter and returns the corresponding
// DownloadInfoDTO as JSON. If the parameter is missing, it responds with
// HTTP 400. If no matching record is found, it responds with HTTP 404.
// Unexpected errors result in HTTP 500.
//
// Swagger annotations:
//
// @Summary      Get DownloadInfo by PCBANumber
// @Description  Returns the DownloadInfo for the specified PCBANumber
// @Tags         download
// @Accept       json
// @Produce      json
// @Param        pcbanumber  query     string  true  "PCBA Number"
// @Success      200  {object}  dto.DownloadInfoDTO
// @Failure      400  {object}  map[string]string  "pcbanumber is required"
// @Failure      404  {object}  map[string]string  "not found"
// @Failure      500  {object}  map[string]string  "internal error"
// @Router       /download [get]
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
