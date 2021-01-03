package db

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

//NumToAttr ...
func NumToAttr(val int) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(int64(val), 10))}
}

//AttrToNum ...
func AttrToNum(attrVal *dynamodb.AttributeValue) (int, error) {
	val, err := strconv.ParseInt(*attrVal.N, 10, 64)
	return int(val), err
}

//Num64ToAttr ...
func Num64ToAttr(val int64) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(val, 10))}
}

//AttrToNum64 ...
func AttrToNum64(attrVal *dynamodb.AttributeValue) (int64, error) {
	return strconv.ParseInt(*attrVal.N, 10, 64)
}

//StrToAttr ...
func StrToAttr(val string) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{S: aws.String(val)}
}

//AttrToStr ...
func AttrToStr(attrVal *dynamodb.AttributeValue) string {
	return *attrVal.S
}

//StrSetToAttr ....
func StrSetToAttr(val []string) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{SS: aws.StringSlice(val)}
}

//AttrToStrSet ...
func AttrToStrSet(attrVal *dynamodb.AttributeValue) []string {
	return aws.StringValueSlice(attrVal.SS)
}
