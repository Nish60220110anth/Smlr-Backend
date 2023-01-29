package index

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func Get(ctx *gin.Context) {
	time.Sleep(time.Duration(3) * time.Second)
	ctx.String(http.StatusOK, "Oh! i got your message and this is your response")
}
