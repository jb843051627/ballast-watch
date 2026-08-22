package service
import("context";"testing")
func TestBug14_DashboardHonorsCancel(t *testing.T){svc:=ballastTestServices(t);ctx,cancel:=context.WithCancel(ballastCtx());cancel();if err:=svc.Dashboard.Refresh(ctx);err==nil{t.Fatal("cancelled dashboard refresh returned nil")}}
