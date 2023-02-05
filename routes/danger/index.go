package danger

import (
	"github.com/gin-gonic/gin"
	"go_backend/routes/userinfo"
	"go_backend/util"
	"log"
	"net/http"
)

const FILENAME = "danger/index.go"

func CheckError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

var workerInfoList []*userinfo.WorkerInfo

func init() {
	workerInfoList = make([]*userinfo.WorkerInfo, 10)
}

func Get(ctx *gin.Context) {
	util.DebugPrint(FILENAME, "GET", "Just Test")
	ctx.JSON(http.StatusOK, workerInfoList)
	workerInfoList = workerInfoList[:0]
}

func Post(ctx *gin.Context) {
	util.DebugPrint(FILENAME, "POST", "Just Test")
	rawWorkInfo := userinfo.RawWorkerInfo{}
	err := ctx.BindJSON(&rawWorkInfo)
	CheckError(err)
	workerInfoList = append(workerInfoList, rawWorkInfo.ConvertToWorkInfo())
}
