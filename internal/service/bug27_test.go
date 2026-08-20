package service
import("testing";"ballast-watch/internal/model")
func TestBug27_VesselGradeValidation(t *testing.T){svc:=ballastTestServices(t);v:=mustVessel(t,svc);updated,err:=svc.Vessels.Update(ballastCtx(),v.ID,&model.VesselInput{Name:"ship",Code:"VES-01",Grade:"INVALID",AreaSqm:100});if err==nil||updated!=nil{t.Fatalf("invalid grade accepted: %v",err)}}
