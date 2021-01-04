package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jeromelesaux/greenserver/persistence"
)

type Aps struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Body     string `json:"body"`
}

type NotificationMessage struct {
	BundleId    string `json:"bundleID"`
	DeviceToken string `json:"deviceToken"`
	Type        string `json:"type"`
	Aps         Aps    `json:"alert"`
}

type DeviceToken struct {
	DeviceToken string `json:"deviceToken"`
}

type Controller struct {
}

func (ctr *Controller) Healthy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
	return
}

func (ctr *Controller) RegisterDevice(c *gin.Context) {
	var notif NotificationMessage
	err := c.BindJSON(&notif)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	var emptyAps = Aps{}
	// test mandatories fields
	if strings.ToUpper(notif.Type) == "ALERT" && notif.Aps == emptyAps {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "alert must have an aps message",
		})
		return
	}

	d, err := persistence.GetDeviceByToken(notif.DeviceToken)
	if err == nil {
		msg := fmt.Sprintf("Device token [%s] already registered with uid [%s], skip step.\n", notif.DeviceToken, d.Uid)
		fmt.Fprintf(os.Stderr, msg)
		c.JSON(http.StatusAlreadyReported, gin.H{
			"token":   notif.DeviceToken,
			"uid":     d.Uid,
			"message": msg,
		})
		return
	}

	aps, err := json.Marshal(notif.Aps)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"token":   notif.DeviceToken,
			"message": "alert must have an aps message",
		})
		return
	}
	if err := persistence.InsertNewDevice(persistence.NewDeviceTable(notif.DeviceToken, notif.BundleId, notif.Type, string(aps))); err != nil {
		msg := fmt.Sprintf("Error while persisting new token [%s] with error %v\n", notif.DeviceToken, err)
		fmt.Fprintf(os.Stderr, msg)
		c.JSON(http.StatusInternalServerError, gin.H{
			"token":   notif.DeviceToken,
			"uid":     "",
			"message": msg,
		})
		return
	}
	d, err = persistence.GetDeviceByToken(notif.DeviceToken)
	if err == nil {
		msg := fmt.Sprintf("Device token [%s] registered with uid [%s], skip step.\n", notif.DeviceToken, d.Uid)
		fmt.Fprintf(os.Stderr, msg)
		c.JSON(http.StatusOK, gin.H{
			"token":   notif.DeviceToken,
			"uid":     d.Uid,
			"message": msg,
		})
		return
	}
	msg := fmt.Sprintf("Error while getting token [%s] with error %v\n", notif.DeviceToken, err)
	fmt.Fprintf(os.Stderr, msg)
	c.JSON(http.StatusInternalServerError, gin.H{
		"token":   notif.DeviceToken,
		"uid":     "",
		"message": msg,
	})
	return
}
