package policy

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *PolicyController) Delete(c *gin.Context) {
	if err := p.srv.Policy().Delete(c, c.Param("ip")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 5,
			"msg":  err,
			"data": nil,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 6,
		"msg":  "response from Policy delete",
		"data": nil,
	})
}
