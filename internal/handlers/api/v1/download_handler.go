package v1

import (
	"log"
	"log-parser/internal/services/downloadinfo"
	"net/http"
)

type DownloadHandler struct {
	svc downloadinfo.DownloadInfoService
}

func NewDownloadHandler(svc downloadinfo.DownloadInfoService) *DownloadHandler {
	return &DownloadHandler{svc: svc}
}

// Get godoc
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
