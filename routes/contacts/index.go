package contacts

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

const FILENAME = "contact/index.go"
const TABLENAME string = "Contact"

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

type Contact struct {
	Name          string // Secondary Key
	PhoneNumber   string // Prime Key
	Specification string
}

func (ctt *Contact) GetKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: ctt.PhoneNumber}}
}

func (tClient *TClientUserInfo) GetAllContactInfo() []Contact {
	var contactList []Contact
	var err error
	var response *dynamodb.ScanOutput
	projEx := expression.NamesList(
		expression.Name("Name"), expression.Name("PhoneNumber"), expression.Name("Specification"))
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

func (tClient *TClientUserInfo) DeleteContact(info *Contact) error {
	_, err := tClient.DynamoDbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(tClient.TableName), Key: info.GetKey(),
	})
	if err != nil {
		log.Printf("Couldn't delete %v from the table. Here's why: %v\n", info, err)
	}
	return err
}

func (tClient *TClientUserInfo) InsertContact(workerInfo *Contact) error {
	item, err := attributevalue.MarshalMap(workerInfo)
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
	contactList := tClient.GetAllContactInfo()
	ctx.JSON(http.StatusOK, contactList)
}

func Post(ctx *gin.Context) {
	util.DebugPrint(FILENAME, "POST", "Just Test")
	contact := Contact{}
	err := ctx.BindJSON(contact)
	CheckError(err)
	err = tClient.InsertContact(&contact)
	CheckError(err)
}

func Delete(ctx *gin.Context) {
	id, isFound := ctx.GetQuery("PhoneNumber")

	if isFound {
		err := tClient.DeleteContact(&Contact{PhoneNumber: id})
		CheckError(err)
	} else {
		ctx.String(http.StatusBadRequest, "PhoneNumber not provided")
	}
}
