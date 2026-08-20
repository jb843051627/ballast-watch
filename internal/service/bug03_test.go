package service
import("sync";"testing";"ballast-watch/internal/store")
func TestBug03_CacheConcurrentSetGet(t *testing.T){ c:=store.NewCache(); var wg sync.WaitGroup; wg.Add(2); go func(){defer wg.Done();for i:=0;i<20000;i++{c.Set(1,&store.TankSnapshot{BallastTankID:1})}}();go func(){defer wg.Done();for i:=0;i<20000;i++{_ = c.GetAll()}}();wg.Wait() }
