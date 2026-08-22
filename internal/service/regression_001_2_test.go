package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"ballast-watch/internal/config"
	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
)

func ballastTestServices(t *testing.T) *Services {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "ballast.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewServices(db, config.New())
}

func ballastCtx() context.Context { return context.Background() }

func mustVessel(t *testing.T, svc *Services) *model.Vessel {
	t.Helper()
	v, err := svc.Vessels.Create(ballastCtx(), &model.VesselInput{Name: "远洋测试船", Code: "VES-01", Grade: "ISO7", AreaSqm: 100})
	if err != nil {
		t.Fatalf("create vessel: %v", err)
	}
	return v
}

func mustTank(t *testing.T, svc *Services, vessel *model.Vessel) *model.BallastTank {
	t.Helper()
	tank, err := svc.BallastTanks.Create(ballastCtx(), &model.TankInput{VesselID: vessel.ID, Name: "压载舱 A", Code: "TANK-A", Kind: "atelier", AreaSqm: 80, TargetPressure: 25, PressureTolerance: 5, TargetTemp: 22, TargetHumidity: 50})
	if err != nil {
		t.Fatalf("create tank: %v", err)
	}
	return tank
}

func mustPointBallast(t *testing.T, svc *Services, tank *model.BallastTank) *model.SamplingPoint {
	t.Helper()
	point, err := svc.Points.Create(ballastCtx(), &model.SamplingPointInput{BallastTankID: tank.ID, Code: "SP-SAL", ParamType: model.ParamHumidity, ThresholdMin: 10, ThresholdMax: 80, AlarmDurationSec: 5, Enabled: true})
	if err != nil {
		t.Fatalf("create sampling point: %v", err)
	}
	return point
}

func mustSensorBallast(t *testing.T, svc *Services, point *model.SamplingPoint) *model.Sensor {
	t.Helper()
	sensor, err := svc.Sensors.Register(ballastCtx(), &model.SensorInput{SamplingPointID: point.ID, Serial: "BW-SENSOR-01", Vendor: "marine-lab", Battery: 90, CalibrationDueAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatalf("register sensor: %v", err)
	}
	return sensor
}