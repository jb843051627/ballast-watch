package service
import (
 "testing"
 "ballast-watch/internal/model"
)
func TestBug01_UnknownSamplingPointRejected(t *testing.T) {
 svc:=ballastTestServices(t)
 _,err:=svc.WaterReadings.Ingest(ballastCtx(),&model.WaterWaterReadingTreatmentCycle{WaterReadings:[]model.WaterWaterReadingInput{{SamplingPointID:9999,ParamType:model.ParamHumidity,Value:40,MeasuredAt:"2026-01-01T00:00:00Z"}}})
 if err==nil { t.Fatal("unknown sampling point must be rejected") }
}
