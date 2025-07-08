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

func (h *TestStationHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pcba := r.URL.Query().Get("pcbanumber")
	if pcba == "" {
		respondError(w, http.StatusBadRequest, "pcbanumber is required")
		return
	}

	// Get DTOs for response
	records, err := h.testStationSvc.GetByPCBANumber(ctx, pcba)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch records")
		return
	}

	// Get DB records with IDs for fetching steps
	dbRecords, err := h.testStationSvc.GetDbObjectsByPCBANumber(ctx, pcba)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch DB records")
		return
	}

	// Build response
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

		// Use matching DB record ID
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
