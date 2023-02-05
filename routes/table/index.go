/*
Table Package Helps to do operations in table
Presently supported operations are create,delete
*/

package table

import (
	"context"
	"fmt"
	"go_backend/util"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
)

var client *dynamodb.Client

func init() {
	client = GetClientFromEnv()
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

type TableOperation struct {
	Name        string // Table Name
	Operation   string
	PrimKeyName string
	SecKeyName  string
}

type TableOperationList struct {
	tableList []TableOperation
}

var tableOpsList []TableOperation

const FILENAME string = "table/index.go"

func AddTableOperation(tableOps TableOperation) {
	tableOpsList = append(tableOpsList, tableOps)
	util.DebugPrint(FILENAME, "AddTableOperation", "Added to TableOperation")
}

func Get(ctx *gin.Context) {
	ctx.JSON(200, TableOperationList{tableList: tableOpsList})
	util.DebugPrint(FILENAME, "Get", "Get Request finished")
}

// Post Accepts Table Operation , creates and delete tables
/*
  sent : {
   	Name        string  // Table Name
	Operation   string  // CREATE or DELETE
	PrimKeyName string  // Set only if Operation=CREATE
	SecKeyName  string  // same as above
}

*/
func Post(ctx *gin.Context) {
	tableOps := &TableOperation{}
	err := ctx.BindJSON(tableOps)
	CheckError(err)
	util.DebugPrint(FILENAME, "POST", fmt.Sprintf("%v", tableOps))

	if tableOps.Operation == "CREATE" {

		_, err := client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
			AttributeDefinitions: []types.AttributeDefinition{{
				AttributeName: aws.String(tableOps.PrimKeyName),
				AttributeType: types.ScalarAttributeTypeS,
			}, {
				AttributeName: aws.String(tableOps.SecKeyName),
				AttributeType: types.ScalarAttributeTypeS,
			}},
			KeySchema: []types.KeySchemaElement{{
				AttributeName: aws.String(tableOps.PrimKeyName),
				KeyType:       types.KeyTypeHash,
			}, {
				AttributeName: aws.String(tableOps.SecKeyName),
				KeyType:       types.KeyTypeRange,
			}},
			TableName: aws.String(tableOps.Name),
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(5),
				WriteCapacityUnits: aws.Int64(5),
			},
		})

		if err != nil {
			log.Printf("Couldn't create table %v. Here's why: %v\n", tableOps.Name, err)
		} else {
			waiter := dynamodb.NewTableExistsWaiter(client)
			err = waiter.Wait(context.Background(), &dynamodb.DescribeTableInput{
				TableName: aws.String(tableOps.Name)}, 1*time.Minute)
			if err != nil {
				log.Printf("Wait for table exists failed. Here's why: %v\n", err)
			}
			util.DebugPrint(FILENAME, "POST", "Created Table")
		}
	} else if tableOps.Operation == "DELETE" {
		_, err := client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
			TableName: aws.String(tableOps.Name)})
		CheckError(err)
		util.DebugPrint(FILENAME, "POST", "Deleted Table")
	}
}
