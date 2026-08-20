package service
import("context";"testing";"time")
func TestBug20_SummaryHonorsCancel(t *testing.T){svc:=ballastTestServices(t);ctx,cancel:=context.WithCancel(ballastCtx());cancel();if _,err:=svc.Reports.Summary(ctx,time.Time{},time.Time{});err==nil{t.Fatal("summary ignored cancellation")}}
