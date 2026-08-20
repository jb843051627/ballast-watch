package service

import (
	"ballast-watch/internal/alerter"
	"ballast-watch/internal/config"
	"ballast-watch/internal/store"
)

// Services 聚合所有业务服务与依赖。
type Services struct {
	Vessels          *VesselService
	BallastTanks     *TankService
	Points           *SamplingPointService
	WaterReadings    *WaterWaterReadingService
	ComplianceAlerts *ComplianceComplianceAlertService
	Rules            *ComplianceRuleService
	Sensors          *SensorService
	Calibrations     *CalibrationService
	TreatmentCyclees *TreatmentCycleService
	Reports          *ReportService
	Dashboard        *DashboardService
	Export           *ExportService

	Store  *store.DB
	Cache  *store.Cache
	Config *config.AppConfig
	Engine *compliance_alerter.Engine
}

// NewServices 组装全部服务。
func NewServices(db *store.DB, cfg *config.AppConfig) *Services {
	vesselStore := store.NewVesselStore(db)
	tankStore := store.NewTankStore(db)
	sampling_pointStore := store.NewSamplingPointStore(db)
	sensorStore := store.NewSensorStore(db)
	water_readingStore := store.NewWaterWaterReadingStore(db)
	compliance_alertStore := store.NewComplianceComplianceAlertStore(db)
	ruleStore := store.NewComplianceRuleStore(db)
	calStore := store.NewCalibrationStore(db)
	cycleStore := store.NewTreatmentCycleStore(db)
	statusStore := store.NewComplianceStatusStore(db)

	cache := store.NewCache()
	engine := compliance_alerter.NewEngine(ruleStore, compliance_alertStore, sampling_pointStore)

	return &Services{
		Vessels:          NewVesselService(vesselStore),
		BallastTanks:     NewTankService(tankStore, vesselStore),
		Points:           NewSamplingPointService(sampling_pointStore, tankStore),
		WaterReadings:    NewWaterWaterReadingService(water_readingStore, sampling_pointStore, sensorStore, compliance_alertStore, cache, engine),
		ComplianceAlerts: NewComplianceComplianceAlertService(compliance_alertStore),
		Rules:            NewComplianceRuleService(ruleStore, tankStore),
		Sensors:          NewSensorService(sensorStore, sampling_pointStore),
		Calibrations:     NewCalibrationService(calStore, sensorStore),
		TreatmentCyclees: NewTreatmentCycleService(cycleStore, tankStore, statusStore, compliance_alertStore),
		Reports:          NewReportService(water_readingStore, sampling_pointStore, tankStore, vesselStore, compliance_alertStore, statusStore),
		Dashboard:        NewDashboardService(tankStore, sampling_pointStore, cache, compliance_alertStore, water_readingStore),
		Export:           NewExportService(water_readingStore, sampling_pointStore, tankStore),

		Store:  db,
		Cache:  cache,
		Config: cfg,
		Engine: engine,
	}
}
