package service
import("testing";"time";"ballast-watch/internal/model")
 func TestBug06_TrendUnknownTankReturnsError(t *testing.T){svc:=ballastTestServices(t);_,err:=svc.Reports.Trend(ballastCtx(),9999,model.ParamHumidity,time.Now().Add(-time.Hour),time.Now(),10);if err==nil{t.Fatal("unknown tank must return error")}}
