package service
import("context";"errors";"testing";"ballast-watch/internal/model";"ballast-watch/internal/store")
type failingSamplingStore16 struct{store.SamplingPointStore}
func(f failingSamplingStore16)ListByBallastTank(context.Context,int64)([]*model.SamplingPoint,error){return nil,errors.New("sampling query failed")}
func TestBug16_DashboardPropagatesSamplingError(t *testing.T){svc:=ballastTestServices(t);v:=mustVessel(t,svc);tank:=mustTank(t,svc,v);d:=NewDashboardService(store.NewTankStore(svc.Store),failingSamplingStore16{store.NewSamplingPointStore(svc.Store)},store.NewCache(),store.NewComplianceComplianceAlertStore(svc.Store),store.NewWaterWaterReadingStore(svc.Store));_ = tank;if err:=d.Refresh(ballastCtx());err==nil{t.Fatal("dashboard swallowed sampling error")}}
