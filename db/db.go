package db

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	//HKeyName ...
	HKeyName = "HKey"

	//RKeyName ...
	RKeyName = "RKey"
)

var dbapi *dynamodb.DynamoDB

//InitDBAPI ...
func InitDBAPI(region, endpoint, accesskey, secretkey string) {
	if dbapi != nil {
		return
	}

	creds := credentials.NewStaticCredentials(accesskey, secretkey, "")
	cfg := &aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
		Endpoint:    aws.String(endpoint),
	}
	sess := session.Must(session.NewSession())
	dbapi = dynamodb.New(sess, cfg)
}

//DoesTableExit ...
func DoesTableExit(tableName string) (bool, error) {
	_, err := dbapi.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if !ok {
			return false, err
		}
		if aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
			return false, nil
		}
	}
	return true, nil
}

//CreateTable ...
func CreateTable(tableName string, readCapacityUnits, writeCapacityUnits int64) error {

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(HKeyName),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(RKeyName),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(HKeyName),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(RKeyName),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(readCapacityUnits),
			WriteCapacityUnits: aws.Int64(writeCapacityUnits),
		},
		TableName: aws.String(tableName),
	}

	_, err := dbapi.CreateTable(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeResourceInUseException || aerr.Code() == dynamodb.ErrCodeTableAlreadyExistsException || aerr.Code() == dynamodb.ErrCodeTableInUseException {
				fmt.Printf("Table: %v already exists\n", tableName)
				return nil
			}
		} else {
			fmt.Printf("CreateTable Failed: %v", err)
			return err
		}
	}

	fmt.Println("Created the table", tableName)
	return nil
}

//Put ...
//TODO: Improvise to take Item only
func Put(tableName string, item map[string]*dynamodb.AttributeValue) error {
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	}
	_, err := dbapi.PutItem(input)
	if err != nil {
		return err
	}
	return nil
}

//Get ...
func Get(tableName string, key map[string]*dynamodb.AttributeValue) (map[string]*dynamodb.AttributeValue, error) {
	input := &dynamodb.GetItemInput{
		TableName:      aws.String(tableName),
		Key:            key,
		ConsistentRead: aws.Bool(true),
	}
	getRes, err := dbapi.GetItem(input)
	if err != nil {
		return nil, err
	}
	return getRes.Item, nil
}

//Increment ...
func Increment(tableName string, key map[string]*dynamodb.AttributeValue, attr string, incrementBy int) (int, error) {
	val := NumToAttr(incrementBy)
	attrUpdate := &dynamodb.AttributeValueUpdate{Action: aws.String("ADD"), Value: val}
	upadtes := map[string]*dynamodb.AttributeValueUpdate{attr: attrUpdate}

	updateInput := &dynamodb.UpdateItemInput{
		TableName:        aws.String(tableName),
		Key:              key,
		AttributeUpdates: upadtes,
		ReturnValues:     aws.String("UPDATED_NEW"),
	}
	updateRes, err := dbapi.UpdateItem(updateInput)
	if err != nil {
		return -1, err
	}

	retVal, err := AttrToNum(updateRes.Attributes[attr])
	if err != nil {
		return -1, err
	}
	return retVal, nil
}

//Query ...
func Query(tableName string, hkeyVal string, filter map[string]*dynamodb.Condition) (res []map[string]*dynamodb.AttributeValue, err error) {
	keyCond := map[string]*dynamodb.Condition{
		HKeyName: &dynamodb.Condition{
			ComparisonOperator: aws.String("EQ"),
			AttributeValueList: []*dynamodb.AttributeValue{StrToAttr(hkeyVal)},
		},
	}

	input := &dynamodb.QueryInput{
		TableName:      aws.String(tableName),
		ConsistentRead: aws.Bool(true),
		KeyConditions:  keyCond,
		QueryFilter:    filter,
	}

	op, err := dbapi.Query(input)
	if err != nil {
		return nil, err
	}
	return op.Items, nil
}

//Update ...
func Update(tableName string, key map[string]*dynamodb.AttributeValue, updateInfo map[string]*dynamodb.AttributeValue) (err error) {
	updates := make(map[string]*dynamodb.AttributeValueUpdate)
	for attr, attrVal := range updateInfo {
		updates[attr] = &dynamodb.AttributeValueUpdate{
			Action: aws.String("PUT"),
			Value:  attrVal,
		}
	}
	input := &dynamodb.UpdateItemInput{
		TableName:        aws.String(tableName),
		Key:              key,
		AttributeUpdates: updates,
	}
	_, err = dbapi.UpdateItem(input)
	return err
}

//Delete ...
func Delete(tableName string, key, cond map[string]*dynamodb.AttributeValue) (err error) {
	var expected map[string]*dynamodb.ExpectedAttributeValue
	if cond != nil {
		expected = make(map[string]*dynamodb.ExpectedAttributeValue)
		for attr, attrVal := range cond {
			expected[attr] = &dynamodb.ExpectedAttributeValue{
				Value: attrVal,
			}
		}
	}
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key:       key,
		Expected:  expected,
	}
	_, err = dbapi.DeleteItem(input)
	return err
}

//UpdateExclusive ....
func UpdateExclusive(tableName string, key map[string]*dynamodb.AttributeValue, updateInfo, cond map[string]*dynamodb.AttributeValue) (err error) {
	updates := make(map[string]*dynamodb.AttributeValueUpdate)
	for attr, attrVal := range updateInfo {
		updates[attr] = &dynamodb.AttributeValueUpdate{
			Action: aws.String("PUT"),
			Value:  attrVal,
		}
	}
	var expected map[string]*dynamodb.ExpectedAttributeValue
	if cond != nil {
		expected = make(map[string]*dynamodb.ExpectedAttributeValue)
		for attr, attrVal := range cond {
			expected[attr] = &dynamodb.ExpectedAttributeValue{
				Value: attrVal,
			}
		}
	}
	input := &dynamodb.UpdateItemInput{
		TableName:        aws.String(tableName),
		Key:              key,
		Expected:         expected,
		AttributeUpdates: updates,
	}
	_, err = dbapi.UpdateItem(input)
	return err

}
