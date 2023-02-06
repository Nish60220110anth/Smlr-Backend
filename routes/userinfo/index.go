package userinfo

import (
	"cloud.google.com/go/civil"
	"context"
	"fmt"
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
	"time"
)

const TABLENAME string = "WorkerInfo"

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

type RawWorkerInfo struct {
	GroundNumber string
	HelmetNumber string
	Spo2Level    int16
	Temperature  int32
	GasLevel     int32
	HeartRate    int32
	DangerType   string // Only two types possible (SOS ,Water)
}

type WorkerInfo struct {
	Id            string // concat of groundNum and HelmetNum, Use this as unique identifier to get specific person identity
	Name          string // This field must be manually Updated at end of the day ( until then a random value is sent)
	Spo2Level     int16
	Temperature   int32
	GasLevel      int32
	HeartRate     int32
	DangerType    string // Only two types possible (SOS ,Water)
	Date          string // Secondary Index
	TreatedDoctor string
}

func (rawWorkInfo *RawWorkerInfo) ConvertToWorkInfo() *WorkerInfo {
	workInfo := &WorkerInfo{}
	workInfo.FillRawFields(rawWorkInfo)
	workInfo.Name = "Just-xxx"
	workInfo.TreatedDoctor = "Just-Doc-XXX"
	workInfo.Date = civil.DateOf(time.Now()).String()

	return workInfo
}

func (wInfo *WorkerInfo) GetKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: wInfo.Id}, "Date": &types.AttributeValueMemberS{Value: wInfo.Date}}
}

func (workerInfo *WorkerInfo) FillRawFields(ruserInfo *RawWorkerInfo) {
	workerInfo.Id = ruserInfo.GroundNumber + "_" + ruserInfo.HelmetNumber
	workerInfo.Spo2Level = ruserInfo.Spo2Level
	workerInfo.Temperature = ruserInfo.Temperature
	workerInfo.GasLevel = ruserInfo.GasLevel
	workerInfo.HeartRate = ruserInfo.HeartRate
	workerInfo.DangerType = ruserInfo.DangerType
}

const FILENAME = "userinfo/index.go"

func (tClient *TClientUserInfo) GetWorkerInfo(id string) WorkerInfo {
	workerInfo := WorkerInfo{Id: id}
	response, err := tClient.DynamoDbClient.GetItem(context.Background(), &dynamodb.GetItemInput{
		Key: workerInfo.GetKey(), TableName: aws.String(tClient.TableName),
	})
	if err != nil {
		log.Printf("Couldn't get info about %v. Reason => : %v\n", id, err)
	} else {
		err = attributevalue.UnmarshalMap(response.Item, &workerInfo)
		if err != nil {
			log.Printf("Couldn't unmarshal response. Reason => : %v\n", err)
		}
	}
	return workerInfo
}

// GetAllWorkerInfo Returns all worker info recorded btw start and end date
func (tClient *TClientUserInfo) GetAllWorkerInfo(startDate, endDate string) []WorkerInfo {
	var workerInfoList []WorkerInfo
	var err error
	var response *dynamodb.ScanOutput
	filtExpre := expression.Name("Date").Between(expression.Value(startDate), expression.Value(endDate))
	projEx := expression.NamesList(
		expression.Name("Id"), expression.Name("Date"), expression.Name("TreatedDoctor"),
		expression.Name("Name"), expression.Name("Spo2Level"), expression.Name("GasLevel"),
		expression.Name("Temperature"), expression.Name("DangerType"), expression.Name("HeartRate"))
	expr, err := expression.NewBuilder().WithFilter(filtExpre).WithProjection(projEx).Build()
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
			log.Printf("Couldn't scan for workerinfo between %v and %v. Here's why: %v\n",
				startDate, endDate, err)
		} else {
			err = attributevalue.UnmarshalListOfMaps(response.Items, &workerInfoList)
			if err != nil {
				log.Printf("Couldn't unmarshal query response. Here's why: %v\n", err)
			}
		}
	}
	return workerInfoList
}

func (tClient *TClientUserInfo) InsertWorkerInfo(workerInfo *WorkerInfo) error {
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

func (tClient *TClientUserInfo) GetWorkerInfoOnQuery() {

}

func (tClient *TClientUserInfo) DeleteWorkerInfo(info WorkerInfo) error {
	_, err := tClient.DynamoDbClient.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: aws.String(tClient.TableName), Key: info.GetKey(),
	})
	if err != nil {
		log.Printf("Couldn't delete %v from the table. Here's why: %v\n", info, err)
	}
	return err
}

func (tClient *TClientUserInfo) UpdateWorkerInfo(workerInfo *WorkerInfo) (map[string]interface{}, error) {
	var err error
	var response *dynamodb.UpdateItemOutput
	var attributeMap map[string]interface{}

	fmt.Println("Update Called")

	update := expression.Set(expression.Name("Name"), expression.Value(workerInfo.Name))
	update.Set(expression.Name("Spo2Level"), expression.Value(workerInfo.Spo2Level))
	update.Set(expression.Name("Temperature"), expression.Value(workerInfo.Temperature))
	update.Set(expression.Name("GasLevel"), expression.Value(workerInfo.GasLevel))
	update.Set(expression.Name("HeartRate"), expression.Value(workerInfo.HeartRate))
	update.Set(expression.Name("DangerType"), expression.Value(workerInfo.DangerType))
	update.Set(expression.Name("TreatedDoctor"), expression.Value(workerInfo.TreatedDoctor))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		log.Printf("Couldn't build expression for update. Here's why: %v\n", err)
	} else {
		response, err = tClient.DynamoDbClient.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
			TableName:                 aws.String(tClient.TableName),
			Key:                       workerInfo.GetKey(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			UpdateExpression:          expr.Update(),
			ReturnValues:              types.ReturnValueUpdatedNew,
		})
		if err != nil {
			log.Printf("Couldn't update workerInfo %v. Reason => %v\n", *workerInfo, err)
		} else {
			err = attributevalue.UnmarshalMap(response.Attributes, &attributeMap)
			if err != nil {
				log.Printf("Couldn't unmarshall update response. Reason => %v\n", err)
			}
		}
	}
	return attributeMap, err
}

// TODO:
func (tClient *TClientUserInfo) InsertBatchWorkerInfo() {
	//var err error
	//var item map[string]types.AttributeValue
	//written := 0
	//batchSize := 25 // DynamoDB allows a maximum batch size of 25 items.
	//start := 0
	//end := start + batchSize
	//for start < maxMovies && start < len(movies) {
	//	var writeReqs []types.WriteRequest
	//	if end > len(movies) {
	//		end = len(movies)
	//	}
	//	for _, movie := range movies[start:end] {
	//		item, err = attributevalue.MarshalMap(movie)
	//		if err != nil {
	//			log.Printf("Couldn't marshal movie %v for batch writing. Here's why: %v\n", movie.Title, err)
	//		} else {
	//			writeReqs = append(
	//				writeReqs,
	//				types.WriteRequest{PutRequest: &types.PutRequest{Item: item}},
	//			)
	//		}
	//	}
	//	_, err = basics.DynamoDbClient.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
	//		RequestItems: map[string][]types.WriteRequest{basics.TableName: writeReqs}})
	//	if err != nil {
	//		log.Printf("Couldn't add a batch of workerInfo to %v. Here's why: %v\n", basics.TableName, err)
	//	} else {
	//		written += len(writeReqs)
	//	}
	//	start = end
	//	end += batchSize
	//}
	//
	//return written, err
}

// YYYY-MM-DD
func Get(ctx *gin.Context) {
	util.DebugPrint(FILENAME, "GET", "Just Test")
	id, isFound := ctx.GetQuery("Id")
	sdate, sfound := ctx.GetQuery("sdate")
	edate, efound := ctx.GetQuery("edate")

	if isFound {
		ctx.JSON(http.StatusOK, tClient.GetWorkerInfo(id))
	} else if !(sfound && efound) {
		ctx.JSON(http.StatusBadRequest, "")
	} else {
		ctx.JSON(http.StatusOK, tClient.GetAllWorkerInfo(sdate, edate))
	}
}

func Post(ctx *gin.Context) {
	util.DebugPrint(FILENAME, "POST", "Just Test")
	rworkInfo := &RawWorkerInfo{}
	err := ctx.BindJSON(rworkInfo)
	CheckError(err)
	workInfo := rworkInfo.ConvertToWorkInfo()
	err = tClient.InsertWorkerInfo(workInfo)
	CheckError(err)
}

func Update(ctx *gin.Context) {
	workInfo := &WorkerInfo{}
	err := ctx.BindJSON(workInfo)
	CheckError(err)
	_, err = tClient.UpdateWorkerInfo(workInfo)
	CheckError(err)
}

func Delete(ctx *gin.Context) {
	id, isFound := ctx.GetQuery("Id")

	if isFound {
		err := tClient.DeleteWorkerInfo(WorkerInfo{Id: id})
		CheckError(err)
	} else {
		ctx.String(http.StatusBadRequest, "Id not provided")
	}
}
