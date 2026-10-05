package service

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service/example"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
)

var ServiceGroupApp = &ServiceGroup{
	SystemServiceGroup:  system.ServiceGroup{},
	ExampleServiceGroup: example.ServiceGroup{},
	NesmaServiceGroup: nesma.ServiceGroup{
		DocumentExportService:  nesma.NewDocumentExportService(),
		UnifiedAnalysisService: nesma.NewUnifiedAnalysisService(),
	},
}

type ServiceGroup struct {
	SystemServiceGroup  system.ServiceGroup
	ExampleServiceGroup example.ServiceGroup
	NesmaServiceGroup   nesma.ServiceGroup
}
