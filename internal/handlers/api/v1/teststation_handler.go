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
	pcba := r.URL.Query().Get("pcbanumber")
	if pcba == "" {
		respondError(w, http.StatusBadRequest, "pcbanumber is required")
		return
	}

	// 1) Берём все записи TestStationRecord с нужным stationType
	records, err := h.testStationSvc.GetByPCBANumber(r.Context(), pcba)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch records")
		return
	}

	// Фильтруем по stationType
	var out []any
	for _, rec := range records {
		if rec.TestStation != h.stationType {
			continue
		}

		// 2) Получаем LogisticData
		logDTO, err := h.logisticSvc.GetByPCBANumber(r.Context(), pcba)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to fetch logistic")
			return
		}
		rec.LogisticData = logDTO

		// 3) Получаем шаги теста
		steps, err := h.testStepSvc.GetByTestStationRecordID(r.Context(), rec.LogisticDataID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to fetch steps")
			return
		}
		// Вложенный JSON-поле — TestSteps
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
