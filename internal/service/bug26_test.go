package service
import("context";"testing";"time")
 func TestBug26_ExportHonorsCancel(t *testing.T){svc:=ballastTestServices(t);v:=mustVessel(t,svc);tank:=mustTank(t,svc,v);ctx,cancel:=context.WithCancel(ballastCtx());cancel();if _,err:=svc.Export.ExportWaterReadingsCSV(ctx,tank.ID,time.Now().Add(-time.Hour),time.Now());err==nil{t.Fatal("export ignored cancellation")}}
