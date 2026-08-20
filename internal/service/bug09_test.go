package service
import("testing";"time";"ballast-watch/internal/model")
func TestBug09_MissingSensorCalibrationError(t *testing.T){svc:=ballastTestServices(t);_,err:=svc.Calibrations.Record(ballastCtx(),&model.CalibrationInput{SensorID:9999,PerformedAt:time.Now(),DueAt:time.Now().Add(time.Hour),Standard:"BWTS",Result:"pass",Operator:"chief"});if err==nil{t.Fatal("missing sensor must return error")}}
