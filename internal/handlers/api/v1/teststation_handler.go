package v1

import (
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/services/logistic"
	"log-parser/internal/services/teststation"
	"log-parser/internal/services/teststep"
	"net/http"
)

type TestStationHandler struct {
	stationType    string
	logisticSvc    logistic.LogisticDataService
	testStationSvc teststation.TestStationService
	testStepSvc    teststep.TestStepService
}

func NewTestStationHandler(stationType string,
	logisticSvc logistic.LogisticDataService,
	testStationSvc teststation.TestStationService,
	testStepSvc teststep.TestStepService,
) *TestStationHandler {
	return &TestStationHandler{stationType, logisticSvc, testStationSvc, testStepSvc}
}

// Get godoc
// @Summary      Get TestStation records by PCBANumber
// @Description  Returns TestStation records and their related logistic and test step data for a specified PCBANumber
// @Tags         teststation
// @Accept       json
// @Produce      json
// @Param        pcbanumber  query     string  true  "PCBA Number"
// @Success      200  {array}  dto.TestStationWithSteps
// @Failure      400  {object}  map[string]string  "pcbanumber is required"
// @Failure      404  {object}  map[string]string  "no matching records"
// @Failure      500  {object}  map[string]string  "internal server error"
func (h *TestStationHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pcba := r.URL.Query().Get("pcbanumber")
	if pcba == "" {
		respondError(w, http.StatusBadRequest, "pcbanumber is required")
		return
	}

	records, err := h.testStationSvc.GetByPCBANumber(ctx, pcba)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch records")
		return
	}

	dbRecords, err := h.testStationSvc.GetDbObjectsByPCBANumber(ctx, pcba)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch DB records")
		return
	}

	var out []any
	for i, rec := range records {
		if rec.TestStation != h.stationType {
			continue
		}

		logDTO, err := h.logisticSvc.GetByPCBANumber(ctx, pcba)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to fetch logistic")
			return
		}
		rec.LogisticData = logDTO

		if i >= len(dbRecords) {
			respondError(w, http.StatusInternalServerError, "record mismatch between DTOs and DBs")
			return
		}
		steps, err := h.testStepSvc.GetByTestStationRecordID(ctx, dbRecords[i].ID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to fetch steps")
			return
		}

		out = append(out, struct {
			dto.TestStationRecordDTO
			TestSteps []dto.TestStepDTO `json:"TestSteps"`
		}{
			rec,
			steps,
		})
	}

	if len(out) == 0 {
		respondError(w, http.StatusNotFound, "no matching records")
		return
	}

	respondJSON(w, http.StatusOK, out)
}

// GetFinal godoc
// @Summary      Get Final TestStation records by PCBANumber
// @Tags         teststation
// @Produce      json
// @Param        pcbanumber  query     string  true  "PCBA Number"
// @Success      200  {array}  dto.TestStationWithSteps
// @Failure      400  {object}  map[string]string  "pcbanumber is required"
// @Failure      404  {object}  map[string]string  "no matching records"
// @Failure      500  {object}  map[string]string  "internal server error"
// @Router       /final [get]
func (h *TestStationHandler) GetFinal(w http.ResponseWriter, r *http.Request) {
	h.stationType = "Final"
	h.Get(w, r)
}

// GetPCBA godoc
// @Summary      Get PCBA TestStation records by PCBANumber
// @Tags         teststation
// @Produce      json
// @Param        pcbanumber  query     string  true  "PCBA Number"
// @Success      200  {array}  dto.TestStationWithSteps
// @Failure      400  {object}  map[string]string  "pcbanumber is required"
// @Failure      404  {object}  map[string]string  "no matching records"
// @Failure      500  {object}  map[string]string  "internal server error"
// @Router       /pcba [get]
func (h *TestStationHandler) GetPCBA(w http.ResponseWriter, r *http.Request) {
	h.stationType = "PCBA"
	h.Get(w, r)
}

// GetPCBANumbers godoc
// @Summary      Get all PCBANumbers for a TestStation type
// @Description  Returns all PCBANumbers available for the configured TestStation type
// @Tags         teststation
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.PCBANumbersResponse
// @Failure      500  {object}  map[string]string  "failed to fetch PCBA numbers"
// @Router       /pcbanumbers [get]
func (h *TestStationHandler) GetPCBANumbers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pcbaNumbers, err := h.testStationSvc.GetAllPCBANumbers(ctx, h.stationType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch PCBA numbers")
		return
	}

	respondJSON(w, http.StatusOK, struct {
		PCBANumbers []string `json:"PCBANumbers"`
	}{
		PCBANumbers: pcbaNumbers,
	})
}
