package loganalysis

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (con *LogAnalysisController) Redirect(c *gin.Context) {
	loc := con.BaseUrl + c.Request.URL.Path
	c.Redirect(http.StatusFound, loc)
}
