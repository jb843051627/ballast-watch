package service
import (
 "testing"
 "ballast-watch/internal/model"
 "ballast-watch/internal/store"
)
func TestBug02_CacheSnapshotIsIndependent(t *testing.T) {
 svc:=ballastTestServices(t)
 svc.Cache.Set(1,&store.TankSnapshot{BallastTankID:1,Realtime:[]model.RealtimeWaterWaterReading{{SamplingPointID:1,Value:12}}})
 got:=svc.Cache.GetAll(); got[0].Realtime[0].Value=99
 again:=svc.Cache.GetAll()
 if again[0].Realtime[0].Value==99 { t.Fatal("cache snapshot was mutated through returned slice") }
}
