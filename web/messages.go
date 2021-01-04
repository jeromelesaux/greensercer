package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jeromelesaux/greenserver/notification"
	"github.com/jeromelesaux/greenserver/persistence"
)

type Alert struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Body     string `json:"body"`
}

type Aps struct {
	Alert Alert `json:"alert"`
}

type NotificationMessage struct {
	BundleId    string `json:"bundleID"`
	DeviceToken string `json:"deviceToken"`
	Type        string `json:"type"`
	Aps         Aps    `json:"aps"`
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

func (ctr *Controller) Notify(c *gin.Context) {
	var notif NotificationMessage
	err := c.BindJSON(&notif)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
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
	amsg := fmt.Sprintf("{\"aps\":%s}", string(aps))
	fmt.Fprintf(os.Stdout, "APS:[%s]\n", amsg)
	err = notification.Notify(notif.DeviceToken, []byte(amsg))
	if err != nil {
		msg := fmt.Sprintf("Device token [%s] registered, cannot be notified error [%v].\n", notif.DeviceToken, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"token":   notif.DeviceToken,
			"message": msg,
		})
		return
	}
	msg := fmt.Sprintf("Device token [%s] has been notified.\n", notif.DeviceToken)
	c.JSON(http.StatusOK, gin.H{
		"token":   notif.DeviceToken,
		"message": msg,
	})
	return
}

func (ctr *Controller) ForceNotify(c *gin.Context) {
	t := time.Now()
	err := notification.NotifAll(t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "All devices are notified",
	})
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
	amsg := fmt.Sprintf("{\"aps\":%s}", string(aps))
	err = notification.Notify(notif.DeviceToken, []byte(amsg))
	if err := persistence.InsertNewDevice(persistence.NewDeviceTable(notif.DeviceToken, notif.BundleId, notif.Type, amsg)); err != nil {
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
