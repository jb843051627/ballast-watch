package service
import("context";"testing")
func TestBug05_EngineHonorsCancel(t *testing.T){svc:=ballastTestServices(t);ctx,cancel:=context.WithCancel(ballastCtx());cancel();if err:=svc.Engine.Evaluate(ctx,nil);err==nil{t.Fatal("cancelled engine returned nil")}}
