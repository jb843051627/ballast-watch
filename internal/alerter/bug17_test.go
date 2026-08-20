package compliance_alerter
import("testing";"ballast-watch/internal/model")
func TestBug17_SubmitQueuesOnce(t *testing.T){e:=&Evaluator{queue:make(chan []*model.WaterReading,4)};e.Submit(nil);if n:=len(e.queue);n!=1{t.Fatalf("expected one queued batch, got %d",n)}}
