package service
import("context";"database/sql";"errors";"testing";"time")
func TestBug07_TransactionReleasesConnection(t *testing.T){svc:=ballastTestServices(t);_ = svc.Store.WithTx(context.Background(),func(tx *sql.Tx)error{return errors.New("failed cycle")});ctx,cancel:=context.WithTimeout(context.Background(),time.Second);defer cancel();if err:=svc.Store.WithTx(ctx,func(*sql.Tx)error{return nil});err!=nil{t.Fatalf("connection remained occupied: %v",err)}}
