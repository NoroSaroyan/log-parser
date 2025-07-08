package v1

import (
	"github.com/go-chi/chi/v5"
	"log-parser/internal/services/downloadinfo"
	"log-parser/internal/services/logistic"
	"log-parser/internal/services/teststation"
	"log-parser/internal/services/teststep"
)

func RegisterAPIV1(r chi.Router,
	downloadSvc downloadinfo.DownloadInfoService,
	logisticSvc logistic.LogisticDataService,
	testStationSvc teststation.TestStationService,
	testStepSvc teststep.TestStepService,
) {
	r.Route("/api/v1", func(r chi.Router) {
		r.With(JSON...).
			Get("/download", NewDownloadHandler(downloadSvc).Get)

		finalH := NewTestStationHandler("Final", logisticSvc, testStationSvc, testStepSvc)
		r.With(JSON...).
			Get("/final", finalH.Get)

		pcbaH := NewTestStationHandler("PCBA", logisticSvc, testStationSvc, testStepSvc)
		r.With(JSON...).
			Get("/pcba", pcbaH.Get)

		pcbaH = NewTestStationHandler("PCBA", logisticSvc, testStationSvc, testStepSvc)
		r.With(JSON...).
			Get("/pcbanumbers", pcbaH.GetPCBANumbers)
	})
}
