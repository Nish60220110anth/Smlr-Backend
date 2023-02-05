package about

import (
	"cloud.google.com/go/civil"
	"fmt"
	"github.com/gin-gonic/gin"
	"go_backend/util"
	"log"
	"net/http"
	"time"
)

const FILENAME = "about/index.go"

type About struct {
	Name        string
	CreatedDate string
	Creators    []string
	Description string
}

var about About

func init() {
	about = CreateAbout()
}

func CheckError(err error) {
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func CreateAbout() About {
	return About{Name: "Smlr", CreatedDate: fmt.Sprintf("%v", civil.DateOf(time.Now())), Creators: []string{
		"Nishanth", "Swami", "Venki", "Chirag", " Arpit", "Ano-1", "Ano-2",
	},
		Description: "Smart Helmet Monitoring Tool"}
}

func Get(ctx *gin.Context) {
	util.DebugPrint(FILENAME, "GET", "Just Test")
	ctx.JSON(http.StatusOK, about)
}

func Post(ctx *gin.Context) {
	util.DebugPrint(FILENAME, "POST", "Just Test")
	err := ctx.BindJSON(&about)
	CheckError(err)
}
