package db

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	accessKey          = "dummy"
	secretKey          = "dummy"
	region             = "us-east-1"
	endpoint           = "http://127.0.0.1:8000"
	tableName          = "TestTable"
	hKey               = "HKey"
	rKey               = "RKey"
	readCapacityUnits  = 100
	writeCapacityUnits = 100
)

func TestInitDBAPI(t *testing.T) {
	t.Log("TestGetDBAPI")

	InitDBAPI(region, endpoint, accessKey, secretKey)
	if dbapi == nil {
		t.Fatal("TestInitDBAPI Failed Initialize session")
		return
	}
}

func TestDoesTableExist(t *testing.T) {
	t.Log("TestCreateTable")
	testTable := "testDoesTableExist"
	exist, err := DoesTableExit(testTable)
	if err != nil {
		t.Fatalf("TestDoesTableExist Failed. Error: %v\n", err)
		return
	}
	if exist {
		t.Fatalf("Table existing, Expected Not to exist")
		return
	}
}

func TestCreateTable(t *testing.T) {
	t.Log("TestCreateTable")

	err := CreateTable(tableName, readCapacityUnits, writeCapacityUnits)
	if err != nil {
		t.Fatalf("TestCreateTable Failed. Error: %v", err)
		return
	}
}

func TestPut(t *testing.T) {
	t.Log("TestPut")

	testRecord := map[string]*dynamodb.AttributeValue{
		HKeyName:   StrToAttr("testHKey/"),
		RKeyName:   StrToAttr("testRKey"),
		"testAttr": NumToAttr(0),
	}

	err := Put(tableName, testRecord)
	if err != nil {
		t.Fatalf("TestPut Failed. Error: %v", err)
		return
	}

	testRecordKey := map[string]*dynamodb.AttributeValue{
		HKeyName: StrToAttr("testHKey/"),
		RKeyName: StrToAttr("testRKey"),
	}

	res, err := Get(tableName, testRecordKey)
	if err != nil {
		t.Fatalf("TestPut Get Failed: Error: %v\n", err)
		return
	}

	if attrVal, ok := res["testAttr"]; ok {
		_, err := AttrToNum(attrVal)
		if err != nil {
			t.Fatalf("TestPut failed parsing attribute: testAttr. Error: %v\n", err)
			return
		}
	} else {
		t.Fatalf("TestPut failed getting attribute: testAttr. Error: %v\n", err)
		return
	}
}

func TestIncrement(t *testing.T) {
	t.Log("TestIncrement")
	testRecordKey := map[string]*dynamodb.AttributeValue{
		HKeyName: StrToAttr("testHKey/"),
		RKeyName: StrToAttr("testRKey"),
	}
	newVal, err := Increment(tableName, testRecordKey, "testAttr", 1)
	if err != nil {
		t.Fatalf("TestIncrement Increment Failed. Err: %v", err)
		return
	}
	if newVal != 1 {
		t.Fatalf("TestIncrement Expected: 1. Actual: %v", newVal)
		return
	}
}

func TestQuery(t *testing.T) {
	t.Log("TestQuery")
	testRecord := map[string]*dynamodb.AttributeValue{
		HKeyName: StrToAttr("TestQuery/"),
		RKeyName: StrToAttr("1"),
		"Data":   StrToAttr("testData1"),
	}
	Put(tableName, testRecord)
	testRecord = map[string]*dynamodb.AttributeValue{
		HKeyName: StrToAttr("TestQuery/"),
		RKeyName: StrToAttr("2"),
		"Data":   StrToAttr("testData2"),
	}
	Put(tableName, testRecord)
	res, err := Query(tableName, "TestQuery/", nil)
	if err != nil {
		t.Fatalf("TestQuery: Query Failed: Error: %v\n", err)
		return
	}

	if len(res) != 2 {
		t.Fatalf("TestQuery: Expected: 2: Actual: %v\n", len(res))
		return
	}
}
