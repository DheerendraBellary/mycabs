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
	cityRecord["Bookings"] = db.Num64ToAttr(int64(0))

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
	historyRec := fmt.Sprintf("%v. State: %v | From Time: %v", 0, stateIdle, time.Now())

	cabRecord := make(map[string]*dynamodb.AttributeValue)
	cabRecord[db.HKeyName] = db.StrToAttr(hkeyValCabs)
	cabRecord[db.RKeyName] = db.StrToAttr(cabID)
	cabRecord["Id"] = db.StrToAttr(cabID)
	cabRecord["Name"] = db.StrToAttr(req.Name)
	cabRecord["Type"] = db.StrToAttr(req.Type)
	cabRecord["CityID"] = db.StrToAttr(req.CityID)
	cabRecord["State"] = db.StrToAttr(stateIdle)
	cabRecord["LastTrip"] = db.Num64ToAttr(curTime) //Used to calculate max idle time since last trip.
	cabRecord["ToCityID"] = db.StrToAttr("")        //A workaround to avoid separate booking record as of now.
	cabRecord["History"] = db.StrSetToAttr([]string{historyRec})

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

	//Cab History
	history := db.AttrToStrSet(cabRecords[cabIdx]["History"])
	histRec := fmt.Sprintf("%v. State: %v | Traveling From: %v to %v | StartTime: %v", len(history), stateOnTrip, req.From, req.To, time.Now())
	history = append(history, histRec)

	//Update the state of the cab in DB
	updateInfo := map[string]*dynamodb.AttributeValue{
		"State":    db.StrToAttr(stateOnTrip),
		"ToCityID": db.StrToAttr(req.To),
		"History":  db.StrSetToAttr(history),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateIdle),
	}

	err = db.UpdateExclusive(tableName, keys, updateInfo, cond)
	if err != nil {
		fmt.Printf("BookCab: db.UpdateExclusive failed. Err: %v\n", err)
		return nil, err
	}

	//Try Udating BookingCount of the City.
	citykeys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCities),
		db.RKeyName: db.StrToAttr(req.From),
	}
	_, err = db.Increment(tableName, citykeys, "Bookings", 1)
	if err != nil {
		//Just log the error and move ahead to return the cab.
		fmt.Printf("BookCab Failed: %v\n", err)
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

	cabRec, err := db.Get(tableName, keys)
	if err != nil {
		fmt.Printf("EndTrip: db.Get Failed. Err: %v\n", err)
		return err
	}

	cityID := req.CityID
	if cityID == "" {
		cityID = db.AttrToStr(cabRec["ToCityID"])
	}

	//Cab History
	history := db.AttrToStrSet(cabRec["History"])
	histRec := fmt.Sprintf("%v. State: %v | Trip Ended In: %v | EndTime: %v", len(history), stateIdle, cityID, time.Now())
	history = append(history, histRec)

	updateInfo := map[string]*dynamodb.AttributeValue{
		"State":    db.StrToAttr(stateIdle),
		"CityID":   db.StrToAttr(cityID),
		"ToCityID": db.StrToAttr(""),
		"LastTrip": db.Num64ToAttr(time.Now().Unix()),
		"History":  db.StrSetToAttr(history),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateOnTrip),
	}

	err = db.UpdateExclusive(tableName, keys, updateInfo, cond)
	return err
}

//DeActivateCab (A force full update of state) ...
func DeActivateCab(req *mycabsapi.DeActivateCabRequest) error {
	keys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCabs),
		db.RKeyName: db.StrToAttr(req.ID),
	}
	cabRec, err := db.Get(tableName, keys)
	if err != nil {
		fmt.Printf("DeActivateCab: db.Get Failed. Err: %v\n", err)
		return err
	}

	//Cab History
	history := db.AttrToStrSet(cabRec["History"])
	histRec := fmt.Sprintf("%v. State: %v | Time: %v", len(history), stateInActive, time.Now())
	history = append(history, histRec)

	updateInfo := map[string]*dynamodb.AttributeValue{
		"State":   db.StrToAttr(stateInActive),
		"History": db.StrSetToAttr(history),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateIdle),
	}

	err = db.UpdateExclusive(tableName, keys, updateInfo, cond)
	return err
}

//ActivateCab (A force full update of state) ...
func ActivateCab(req *mycabsapi.ActivateCabRequest) error {
	keys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCabs),
		db.RKeyName: db.StrToAttr(req.ID),
	}

	cabRec, err := db.Get(tableName, keys)
	if err != nil {
		fmt.Printf("ActivateCab: db.Get Failed. Err: %v\n", err)
		return err
	}

	//Cab History
	history := db.AttrToStrSet(cabRec["History"])
	histRec := fmt.Sprintf("%v. State: %v | Time: %v", len(history), stateIdle, time.Now())
	history = append(history, histRec)

	updateInfo := map[string]*dynamodb.AttributeValue{
		"State":    db.StrToAttr(stateIdle),
		"LastTrip": db.Num64ToAttr(time.Now().Unix()),
		"History":  db.StrSetToAttr(history),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateInActive),
	}

	err = db.UpdateExclusive(tableName, keys, updateInfo, cond)
	return err
}

//ChangeCity (A force full update of City in InActive State) ...
func ChangeCity(req *mycabsapi.ChangeCityRequest) error {
	keys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCabs),
		db.RKeyName: db.StrToAttr(req.CabID),
	}

	cabRec, err := db.Get(tableName, keys)
	if err != nil {
		fmt.Printf("ActivateCab: db.Get Failed. Err: %v\n", err)
		return err
	}
	curCity := db.AttrToStr(cabRec["CityID"])
	//Cab History
	history := db.AttrToStrSet(cabRec["History"])
	histRec := fmt.Sprintf("%v. City Changed From: %v to %v", len(history), curCity, req.CityID)
	history = append(history, histRec)

	updateInfo := map[string]*dynamodb.AttributeValue{
		"CityID":  db.StrToAttr(req.CityID),
		"History": db.StrSetToAttr(history),
	}
	cond := map[string]*dynamodb.AttributeValue{
		"State": db.StrToAttr(stateInActive),
	}

	err = db.UpdateExclusive(tableName, keys, updateInfo, cond)
	return err
}

//DemandedCity ...
func DemandedCity() (*mycabsapi.DemandCityResonse, error) {

	cityRecords, err := db.Query(tableName, hkeyValCities, nil)

	if err != nil {
		fmt.Printf("DemandedCity: db.Query failed. Err: %v\n", err)
		return nil, err
	}
	if len(cityRecords) == 0 {
		fmt.Printf("DemandedCity: No cities found\n")
		return nil, nil
	}

	maxBookings := int64(0)
	cityIdx := 0
	//Flaw - It returns only one in case of clash
	for idx, cityRec := range cityRecords {
		booking, err := db.AttrToNum64(cityRec["Bookings"])
		if err != nil {
			fmt.Printf("DemandedCity: db.AttrToNum64 Failed. Err: %v\n", err)
			return nil, err
		}
		if booking > maxBookings {
			maxBookings = booking
			cityIdx = idx
		}
	}
	city := &mycabsapi.DemandCityResonse{
		CityID:   db.AttrToStr(cityRecords[cityIdx]["Id"]),
		CityName: db.AttrToStr(cityRecords[cityIdx]["Name"]),
	}
	return city, nil
}

//CabHistory (A force full update of state) ...
func CabHistory(req *mycabsapi.CabHistoryRequest) (*mycabsapi.CabHistoryResonse, error) {
	keys := map[string]*dynamodb.AttributeValue{
		db.HKeyName: db.StrToAttr(hkeyValCabs),
		db.RKeyName: db.StrToAttr(req.CabID),
	}

	cabRec, err := db.Get(tableName, keys)
	if err != nil {
		fmt.Printf("CabHistory: db.Get Failed. Err: %v\n", err)
		return nil, err
	}

	//Cab History
	history := db.AttrToStrSet(cabRec["History"])
	cabHistory := &mycabsapi.CabHistoryResonse{
		History: history,
	}
	return cabHistory, nil
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
