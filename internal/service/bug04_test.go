package service
import("errors";"testing";"ballast-watch/internal/model")
func TestBug04_CycleNotFoundErrorChain(t *testing.T){svc:=ballastTestServices(t);_,err:=svc.TreatmentCyclees.Complete(ballastCtx(),9999);if !errors.Is(err,model.ErrNotFound){t.Fatalf("not-found chain lost: %v",err)}}
