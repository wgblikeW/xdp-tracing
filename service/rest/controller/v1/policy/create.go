package policy

import (
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "github.com/p1nant0m/xdp-tracing/pkg/api/v1"
)

func (p *PolicyController) Create(c *gin.Context) {
	var r v1.Policy
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 3,
			"data": nil,
			"msg":  err,
		})

		return
	}

	if err := p.srv.Policy().Create(c, r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 3,
			"data": nil,
			"msg":  err,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 4,
		"data": nil,
		"msg":  "ok " + r.Policy,
	})
}
