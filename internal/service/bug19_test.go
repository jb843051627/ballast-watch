package service
import("testing";"ballast-watch/internal/model")
func TestBug19_CycleStartPropagatesStateFailure(t *testing.T){svc:=ballastTestServices(t);v:=mustVessel(t,svc);tank:=mustTank(t,svc,v);_,err:=svc.Store.Exec("UPDATE tanks SET status='invalid_state' WHERE id=?",tank.ID);if err!=nil{t.Fatal(err)};_,err=svc.TreatmentCyclees.Start(ballastCtx(),&model.TreatmentCycleInput{BallastTankID:tank.ID,Name:"bad",Product:"ballast",Phase:"treat"});if err==nil{t.Fatal("state transition failure was hidden")}}
