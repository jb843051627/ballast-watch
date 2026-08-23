package service
import("testing";"ballast-watch/internal/model")
func TestBug18_DeletedTankCycleReturnsError(t *testing.T){svc:=ballastTestServices(t);v:=mustVessel(t,svc);tank:=mustTank(t,svc,v);c,err:=svc.TreatmentCyclees.Start(ballastCtx(),&model.TreatmentCycleInput{BallastTankID:tank.ID,Name:"cycle",Product:"ballast",Phase:"treat"});if err!=nil{t.Fatal(err)};if _,err=svc.Store.Exec("DELETE FROM tanks WHERE id = ?",tank.ID);err!=nil{t.Fatal(err)};if _,err=svc.TreatmentCyclees.Complete(ballastCtx(),c.ID);err==nil{t.Fatal("deleted tank must return error")}}
