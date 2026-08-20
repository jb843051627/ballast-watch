package compliance_alerter
import("sync";"testing";"ballast-watch/internal/model")
func TestBug23_SubmitCloseRace(t *testing.T){e:=&Evaluator{queue:make(chan []*model.WaterReading,1)};start:=make(chan struct{});var wg sync.WaitGroup;wg.Add(2);go func(){defer wg.Done();<-start;for i:=0;i<1000;i++{e.Submit(nil)}}();go func(){defer wg.Done();<-start;e.Close()}();close(start);wg.Wait()}
