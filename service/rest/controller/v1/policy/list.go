package policy

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *PolicyController) List(c *gin.Context) {
	policies, err := p.srv.Policy().List(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"data": nil,
			"msg":  err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 2,
		"data": policies,
		"msg":  "response from Policies List",
	})
}
