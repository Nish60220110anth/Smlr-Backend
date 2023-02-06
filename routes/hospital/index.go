package hospital

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"go_backend/util"
	"log"
	"net/http"
)

const FILENAME = "hospital/index.go"
const TABLENAME string = "Hospital"

var tClient *TClientUserInfo

type TClientUserInfo struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

func init() {
	tClient = &TClientUserInfo{}
	tClient.TableName = TABLENAME
	tClient.DynamoDbClient = GetClientFromEnv()
}

func GetClientFromEnv() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.Background(), func(l *config.LoadOptions) error {
		return nil
	})

	CheckError(err)

	client := dynamodb.NewFromConfig(cfg)
	return client
}

func CheckError(err error) {
	if err != nil {
		log.Fatalln(err.Error())
	}
}

type Hospital struct {
	Name        string // Secondary Key
	PhoneNumber string // Prime Key
	Address     string
}

func (ctt *Hospital) GetKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: ctt.PhoneNumber}}
}

func (tClient *TClientUserInfo) GetAllHospitalInfo() []Hospital {
	var contactList []Hospital
	var err error
	var response *dynamodb.ScanOutput
	projEx := expression.NamesList(
		expression.Name("Name"), expression.Name("PhoneNumber"), expression.Name("Address"))
	expr, err := expression.NewBuilder().WithProjection(projEx).Build()
	if err != nil {
		log.Printf("Couldn't build expressions for scan. Here's why: %v\n", err)
	} else {
		response, err = tClient.DynamoDbClient.Scan(context.Background(), &dynamodb.ScanInput{
			TableName:                 aws.String(tClient.TableName),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
		})
		if err != nil {
			CheckError(err)
		} else {
			err = attributevalue.UnmarshalListOfMaps(response.Items, &contactList)
			if err != nil {
				log.Printf("Couldn't unmarshal query response. Here's why: %v\n", err)
			}
		}
	}
	return contactList
}

func (tClient *TClientUserInfo) DeleteHospital(info *Hospital) error {
	_, err := tClient.DynamoDbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(tClient.TableName), Key: info.GetKey(),
	})
	if err != nil {
		log.Printf("Couldn't delete %v from the table. Here's why: %v\n", info, err)
	}
	return err
}

func (tClient *TClientUserInfo) InsertContact(hospitalInfo *Hospital) error {
	item, err := attributevalue.MarshalMap(hospitalInfo)
	CheckError(err)
	_, err = tClient.DynamoDbClient.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(tClient.TableName), Item: item,
	})
	if err != nil {
		log.Printf("Couldn't add item to table. Reason => %v\n", err)
	}
	return err
}

func Get(ctx *gin.Context) {
	util.DebugPrint(FILENAME, "GET", "Just Test")
	contactList := tClient.GetAllHospitalInfo()
	ctx.JSON(http.StatusOK, contactList)
}

func Post(ctx *gin.Context) {
	util.DebugPrint(FILENAME, "POST", "Just Test")
	hospital := &Hospital{}
	err := ctx.BindJSON(hospital)
	CheckError(err)
	err = tClient.InsertContact(hospital)
	CheckError(err)
}

func Delete(ctx *gin.Context) {
	id, isFound := ctx.GetQuery("PhoneNumber")

	if isFound {
		err := tClient.DeleteHospital(&Hospital{PhoneNumber: id})
		CheckError(err)
	} else {
		ctx.String(http.StatusBadRequest, "PhoneNumber not provided")
	}
}
