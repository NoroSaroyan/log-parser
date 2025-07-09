package v1

import (
	"github.com/go-chi/chi/v5"
	"log-parser/internal/services/downloadinfo"
	"log-parser/internal/services/logistic"
	"log-parser/internal/services/teststation"
	"log-parser/internal/services/teststep"
)

// RegisterAPIV1 registers all v1 API routes.
//
// @Summary      Register API v1 routes
// @Description  Registers endpoints for download info and test stations (Final, PCBA)
// @Tags         api,v1
func RegisterAPIV1(r chi.Router,
	downloadSvc downloadinfo.DownloadInfoService,
	logisticSvc logistic.LogisticDataService,
	testStationSvc teststation.TestStationService,
	testStepSvc teststep.TestStepService,
) {
	// GET /api/v1/download
	// @Summary      Get download info by PCBA number
	// @Tags         downloadinfo
	// @Produce      json
	// @Param        pcbanumber query string true "PCBA Number"
	// @Success      200 {object} downloadinfo.DTO
	// @Failure      400 {object} map[string]string
	// @Failure      500 {object} map[string]string
	// @Router       /download [get]
	r.With(JSON...).
		Get("/download", NewDownloadHandler(downloadSvc).Get)

	finalH := NewTestStationHandler("Final", logisticSvc, testStationSvc, testStepSvc)
	// GET /api/v1/final
	// @Summary      Get Final TestStation records by PCBA number
	// @Tags         teststation
	// @Produce      json
	// @Param        pcbanumber query string true "PCBA Number"
	// @Success      200 {array} struct{ /* see TestStationHandler.GetFinal */ }
	// @Failure      400 {object} map[string]string
	// @Failure      404 {object} map[string]string
	// @Failure      500 {object} map[string]string
	// @Router       /final [get]
	r.With(JSON...).
		Get("/final", finalH.GetFinal)

	pcbaH := NewTestStationHandler("PCBA", logisticSvc, testStationSvc, testStepSvc)
	// GET /api/v1/pcba
	// @Summary      Get PCBA TestStation records by PCBA number
	// @Tags         teststation
	// @Produce      json
	// @Param        pcbanumber query string true "PCBA Number"
	// @Success      200 {array} struct{ /* see TestStationHandler.GetPCBA */ }
	// @Failure      400 {object} map[string]string
	// @Failure      404 {object} map[string]string
	// @Failure      500 {object} map[string]string
	// @Router       /pcba [get]
	r.With(JSON...).
		Get("/pcba", pcbaH.GetPCBA)

	// GET /api/v1/pcbanumbers
	// @Summary      Get all PCBA numbers
	// @Tags         teststation
	// @Produce      json
	// @Success      200 {object} struct{PCBANumbers []string `json:"PCBANumbers"`}
	// @Failure      500 {object} map[string]string
	// @Router       /pcbanumbers [get]
	r.With(JSON...).
		Get("/pcbanumbers", pcbaH.GetPCBANumbers)
}
