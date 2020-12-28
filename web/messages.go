package web

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jeromelesaux/greenserver/persistence"
)

type DeviceToken struct {
	DeviceToken string `json:"deviceToken"`
}

type Controller struct {
}

func (ctr *Controller) RegisterDevice(c *gin.Context) {
	token := c.Param("token")
	d, err := persistence.GetDeviceByToken(token)
	if err == nil {
		msg := fmt.Sprintf("Device token [%s] already registered with uid [%s], skip step.\n", token, d.Uid)
		fmt.Fprintf(os.Stderr, msg)
		c.JSON(http.StatusAlreadyReported, gin.H{
			"token":   token,
			"uid":     d.Uid,
			"message": msg,
		})
		return
	}
	if err := persistence.InsertNewDevice(persistence.NewDeviceTable(token)); err != nil {
		msg := fmt.Sprintf("Error while persisting new token [%s] with error %v\n", token, err)
		fmt.Fprintf(os.Stderr, msg)
		c.JSON(http.StatusInternalServerError, gin.H{
			"token":   token,
			"uid":     "",
			"message": msg,
		})
		return
	}
	d, err = persistence.GetDeviceByToken(token)
	if err == nil {
		msg := fmt.Sprintf("Device token [%s] registered with uid [%s], skip step.\n", token, d.Uid)
		fmt.Fprintf(os.Stderr, msg)
		c.JSON(http.StatusOK, gin.H{
			"token":   token,
			"uid":     d.Uid,
			"message": msg,
		})
		return
	}
	msg := fmt.Sprintf("Error while getting token [%s] with error %v\n", token, err)
	fmt.Fprintf(os.Stderr, msg)
	c.JSON(http.StatusInternalServerError, gin.H{
		"token":   token,
		"uid":     "",
		"message": msg,
	})
	return
}
