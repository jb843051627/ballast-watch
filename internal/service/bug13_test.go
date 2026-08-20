package service
import("testing";"ballast-watch/internal/model")
func TestBug13_RestrictedTankCanBecomeRelease(t *testing.T){if !model.CanTransition(model.StateRestricted,model.StateRelease){t.Fatal("restricted ballast tank cannot become release")}}
