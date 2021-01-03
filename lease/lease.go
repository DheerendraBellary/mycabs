package lease

import (
	"errors"
	"mycabs/db"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	minGap        = int64(120)
	renewInterval = 90
)

//Lease ...
type Lease struct {
	tableName string
	key       map[string]*dynamodb.AttributeValue
	timeStamp int64
}

//Load ...
func Load(tableName string, key map[string]*dynamodb.AttributeValue) (ls *Lease, err error) {
	rec, err := db.Get(tableName, key)
	if err != nil {
		return nil, err
	}
	if len(rec) == 0 {
		return nil, errors.New("lease.Load: Key not Found")
	}
	if _, ok := rec["Lease"]; !ok {
		return nil, errors.New("lease.Load: Lease Attr not Found")
	}
	leaseTime, err := db.AttrToNum64(rec["Lease"])
	curTime := time.Now().Unix()

	if curTime-leaseTime <= minGap {
		return nil, errors.New("lease.Load: Record Busy")
	}

	updateInfo := map[string]*dynamodb.AttributeValue{
		"Lease": db.Num64ToAttr(curTime),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"Lease": db.Num64ToAttr(leaseTime),
	}

	err = db.UpdateExclusive(tableName, key, updateInfo, cond)
	if err != nil {
		return nil, err
	}

	ls = &Lease{tableName: tableName,
		key:       key,
		timeStamp: curTime}
	return ls, nil
}

//Renew Keeps on renewing the lease until its aborted.
func (ls *Lease) Renew(abort chan int) {
	timer := time.NewTimer(renewInterval * time.Second)
	for {
		select {
		case <-timer.C:
			err := ls.renew()
			if err != nil {
				return
			}
			timer = time.NewTimer(renewInterval * time.Second)
		case <-abort:
			return
		}
	}
}

//Validate ...
func (ls *Lease) Validate() (err error) {
	rec, err := db.Get(ls.tableName, ls.key)
	if err != nil {
		return err
	}
	if len(rec) == 0 {
		return errors.New("lease.Validate: Key not Found")
	}
	leaseTime, err := db.AttrToNum64(rec["Lease"])
	if leaseTime != ls.timeStamp {
		return errors.New("lease.Validate. Falied to match leaseTime and lease.timeStamp")
	}
	return nil
}

//Release ...
func (ls *Lease) Release() (err error) {
	rec, err := db.Get(ls.tableName, ls.key)
	if err != nil {
		return err
	}
	if len(rec) == 0 {
		return errors.New("lease.Release: Key not Found")
	}
	updateInfo := map[string]*dynamodb.AttributeValue{
		"Lease": db.Num64ToAttr(int64(0)),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"Lease": db.Num64ToAttr(ls.timeStamp),
	}

	err = db.UpdateExclusive(ls.tableName, ls.key, updateInfo, cond)
	return err
}

//GetTimeStamp ...
func (ls *Lease) GetTimeStamp() int64 {
	return ls.timeStamp
}

//renew ...
func (ls *Lease) renew() error {
	rec, err := db.Get(ls.tableName, ls.key)
	if err != nil {
		return err
	}
	if len(rec) == 0 {
		return errors.New("lease.renew: Key not Found")
	}

	curTime := time.Now().Unix()

	if curTime-ls.timeStamp > minGap {
		return errors.New("lease.renew: Trying to renew expired lease")
	}

	updateInfo := map[string]*dynamodb.AttributeValue{
		"Lease": db.Num64ToAttr(curTime),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"Lease": db.Num64ToAttr(ls.timeStamp),
	}

	err = db.UpdateExclusive(ls.tableName, ls.key, updateInfo, cond)
	if err != nil {
		return err
	}

	//Update the time stamp here
	ls.timeStamp = curTime

	return nil
}
