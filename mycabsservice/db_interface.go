package mycabsservice

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"mycabs/db"
	"mycabs/lease"
	"mycabs/mycabsapi"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	accessKey          = "dummy"
	secretKey          = "dummy"
	region             = "us-east-1"
	endpoint           = "http://127.0.0.1:8000"
	tableName          = "mycabs"
	hKey               = "HKey"
	rKey               = "RKey"
	readCapacityUnits  = 100
	writeCapacityUnits = 100
)

const (
	hkeyValCabs    = "cabs/"
	hkeyValCities  = "cities/"
	hkeyValCounter = "count/"
)

const (
	stateIdle     = "IDLE"
	stateOnTrip   = "ON_TRIP"
	stateInActive = "IN_ACTIVE"
)

//////////////// Fucntions which are directly called by Service///////////////////////

//OnboardCity ...
func OnboardCity(citiReq *mycabsapi.OnboardCityRequest) (cityID string, err error) {
	//Generate New EmployeeID.
	cityID, err = getNewCityID()
	if err != nil {
		fmt.Printf("OnboardCity: getNewCityID Failed. Error %v\n", err)
		return cityID, err
	}

	cityRecord := make(map[string]*dynamodb.AttributeValue)
	cityRecord[db.HKeyName] = db.StrToAttr(hkeyValCities)
	cityRecord[db.RKeyName] = db.StrToAttr(cityID)
	cityRecord["Id"] = db.StrToAttr(cityID)
	cityRecord["Name"] = db.StrToAttr(citiReq.Name)

	//Store city into DB
	err = db.Put(tableName, cityRecord)
	if err != nil {
		fmt.Printf("OnboardCity: db.Put Failed. Err: %v\n", err)
		return cityID, err
	}

	return cityID, nil
}

//RegisterCab ...
func RegisterCab(req *mycabsapi.RegisterCabRequest) (cabID string, err error) {
	//Generate New EmployeeID.
	cabID, err = getNewCabID()
	if err != nil {
		fmt.Printf("RegisterCab: getNewCabID Failed. Error %v\n", err)
		return cabID, err
	}

	curTime := time.Now().Unix()

	cabRecord := make(map[string]*dynamodb.AttributeValue)
	cabRecord[db.HKeyName] = db.StrToAttr(hkeyValCabs)
	cabRecord[db.RKeyName] = db.StrToAttr(cabID)
	cabRecord["Id"] = db.StrToAttr(cabID)
	cabRecord["Name"] = db.StrToAttr(req.Name)
	cabRecord["Type"] = db.StrToAttr(req.Type)
	cabRecord["CityID"] = db.StrToAttr(req.CityID)
	cabRecord["State"] = db.StrToAttr(stateIdle)
	cabRecord["LastTrip"] = db.Num64ToAttr(curTime) //Used to calculate max idle time since last trip

	//Add the lease value with 0, lease will be used in distributed synchronization.
	//This can be optimized by not setting it now and handling it lease load.
	cabRecord["Lease"] = db.Num64ToAttr(int64(0))

	//Store city into DB
	err = db.Put(tableName, cabRecord)
	if err != nil {
		fmt.Printf("RegisterCab: db.Put Failed. Err: %v\n", err)
		return cabID, err
	}

	return cabID, nil
}

//BookCab ...
func BookCab(req *mycabsapi.BookingRequest) (cab *mycabsapi.Cab, err error) {
	//Bring in the list of cabs which are idle and available in the city.
	//Sort them by idle time and assigns the cab with the most idle time.

	filter := map[string]*dynamodb.Condition{
		"CityID": &dynamodb.Condition{
			ComparisonOperator: aws.String("EQ"),
			AttributeValueList: []*dynamodb.AttributeValue{db.StrToAttr(req.From)},
		},
		"Type": &dynamodb.Condition{
			ComparisonOperator: aws.String("EQ"),
			AttributeValueList: []*dynamodb.AttributeValue{db.StrToAttr(req.CabType)},
		},
		"State": &dynamodb.Condition{
			ComparisonOperator: aws.String("EQ"),
			AttributeValueList: []*dynamodb.AttributeValue{db.StrToAttr(stateIdle)},
		},
	}

	cabRecords, err := db.Query(tableName, hkeyValCabs, filter)

	if err != nil {
		fmt.Printf("BookCab: db.Query failed. Err: %v\n", err)
		return nil, err
	}
	if len(cabRecords) == 0 {
		fmt.Printf("BookCab: No cabs found for the given criteria\n")
		return nil, nil
	}

	cabRecIndex := []int{}
	minLastTripTime := int64(math.MaxInt64)

	for idx, cabRec := range cabRecords {
		lastTripTime, err := db.AttrToNum64(cabRec["LastTrip"])
		if err != nil {
			fmt.Printf("BookCab: db.AttrToNum64 failed. Err: %v\n", err)
			return nil, err
		}
		if lastTripTime < minLastTripTime {
			//Clear the cabRecIndex and store the new
			cabRecIndex = cabRecIndex[:0]
			minLastTripTime = lastTripTime
			cabRecIndex = append(cabRecIndex, idx)
		} else if lastTripTime == minLastTripTime {
			//just add this rec also to the cabRecIndex
			cabRecIndex = append(cabRecIndex, idx)
		}
	}

	//Logic to pick cab with max idle time / radom of them in case of clash.
	cabIdx := 0
	numRecs := len(cabRecIndex)
	if numRecs == 0 {
		//Ideally shouldn't happen
		errmsg := fmt.Sprintf("BookCab: Failed to get the cabs.")
		fmt.Println(errmsg)
		return nil, errors.New(errmsg)
	} else if numRecs == 1 {
		cabIdx = cabRecIndex[0]
	} else {
		cabIdx = cabRecIndex[rand.Intn(numRecs)]
	}

	keys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCabs),
		db.RKeyName: cabRecords[cabIdx]["Id"],
	}

	//Now once the cab is computed, Immeditely take lease on it.
	ls, err := lease.Load(tableName, keys)
	if err != nil {
		//Improvement TODO: There could be a retry mechanism here which can check if there are
		//any other available cabs matching the criteria.

		fmt.Printf("BookCab: lease.Load failed. Err: %v\n", err)
		return nil, err
	}
	abort := make(chan int)
	go ls.Renew(abort)
	defer ls.Release()
	defer close(abort)

	//Update the state of the cab in DB
	updateInfo := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateOnTrip),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateIdle),
	}

	err = db.UpdateExclusive(tableName, keys, updateInfo, cond)
	if err != nil {
		fmt.Printf("BookCab: db.UpdateExclusive failed. Err: %v\n", err)
		return nil, err
	}

	//Now once the state of the cab is changed to ON_TRIP, it is ensured that
	//that is booking is successful, returning it.
	cab = &mycabsapi.Cab{}
	cab.ID = db.AttrToStr(cabRecords[cabIdx]["Id"])
	cab.Name = db.AttrToStr(cabRecords[cabIdx]["Name"])
	cab.Type = db.AttrToStr(cabRecords[cabIdx]["Type"])

	return cab, nil
}

//EndTrip (A force full update of state) ...
func EndTrip(req *mycabsapi.EndTripRequest) error {
	keys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCabs),
		db.RKeyName: db.StrToAttr(req.CabID),
	}
	updateInfo := map[string]*dynamodb.AttributeValue{
		"State":    db.StrToAttr(stateIdle),
		"LastTrip": db.Num64ToAttr(time.Now().Unix()),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateOnTrip),
	}

	err := db.UpdateExclusive(tableName, keys, updateInfo, cond)
	return err
}

//DeActivateCab (A force full update of state) ...
func DeActivateCab(req *mycabsapi.DeActivateCabRequest) error {
	keys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCabs),
		db.RKeyName: db.StrToAttr(req.ID),
	}

	updateInfo := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateInActive),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateIdle),
	}

	err := db.UpdateExclusive(tableName, keys, updateInfo, cond)
	return err
}

//ActivateCab (A force full update of state) ...
func ActivateCab(req *mycabsapi.ActivateCabRequest) error {
	keys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCabs),
		db.RKeyName: db.StrToAttr(req.ID),
	}

	updateInfo := map[string]*dynamodb.AttributeValue{
		"State":    db.StrToAttr(stateIdle),
		"LastTrip": db.Num64ToAttr(time.Now().Unix()),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateInActive),
	}

	err := db.UpdateExclusive(tableName, keys, updateInfo, cond)
	return err
}

//////////////////////////////////////////////////////////////////////////////////////

//getNewCityID : Creates a unique id using Storage Counter and returns
func getNewCityID() (string, error) {
	cityID := ""

	keys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCounter),
		db.RKeyName: db.StrToAttr("city"),
	}

	newCount, err := db.Increment(tableName, keys, "Counter", 1)
	if err != nil {
		fmt.Printf("getNewCityID Failed: %v\n", err)
		return "", err
	}
	cityID = "city_" + strconv.FormatInt(int64(newCount), 10)
	return cityID, nil
}

//getNewCabID : Creates a unique id using Storage Counter and returns
func getNewCabID() (string, error) {
	cabID := ""

	keys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCounter),
		db.RKeyName: db.StrToAttr("cab"),
	}

	newCount, err := db.Increment(tableName, keys, "Counter", 1)
	if err != nil {
		fmt.Printf("getNewCityID Failed: %v\n", err)
		return "", err
	}
	cabID = "cab_" + strconv.FormatInt(int64(newCount), 10)
	return cabID, nil
}

func initCityCounter() error {
	counterRecord := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCounter),
		db.RKeyName: db.StrToAttr("city"),
		"Counter":   db.NumToAttr(0),
	}

	err := db.Put(tableName, counterRecord)
	if err != nil {
		fmt.Printf("initCityCounter: db.Put Failed. Err: %v\n", err)
		return err
	}
	return nil
}

func initCabCounter() error {
	counterRecord := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCounter),
		db.RKeyName: db.StrToAttr("cab"),
		"Counter":   db.NumToAttr(0),
	}

	err := db.Put(tableName, counterRecord)
	if err != nil {
		fmt.Printf("initCabCounter: db.Put Failed. Err: %v\n", err)
		return err
	}
	return nil
}

func init() {
	db.InitDBAPI(region, endpoint, accessKey, secretKey)
	fmt.Println("Initialized DB Session ...")
	exist, err := db.DoesTableExit(tableName)
	if err != nil {
		fmt.Printf("db.DoesTableExit Failed %v\n. Exitting....", err)
		os.Exit(1)
	}
	if exist {
		return
	}
	err = db.CreateTable(tableName, readCapacityUnits, writeCapacityUnits)
	if err != nil {
		fmt.Printf("db.CreateTable Failed %v\n. Exitting....", err)
		os.Exit(1)
	}
	err = initCityCounter()
	if err != nil {
		fmt.Printf("initCityCounter Failed %v\n. Exitting....", err)
		os.Exit(1)
	}
	err = initCabCounter()
	if err != nil {
		fmt.Printf("initCabCounter Failed %v\n. Exitting....", err)
		os.Exit(1)
	}
}
