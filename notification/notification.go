package notification

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jeromelesaux/greenserver/config"
	"github.com/jeromelesaux/greenserver/persistence"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

var (
	GlobalTicker *time.Ticker
	done         = make(chan bool)
)

func Initialise() {
	GlobalTicker = time.NewTicker(1 * time.Hour)
	fmt.Fprintf(os.Stdout, "Initialise ticker each one hour.")
	go func() {
		for t := range GlobalTicker.C {
			if err := NotifAll(t); err != nil {
				fmt.Fprintf(os.Stderr, "Error while trying to notify devices with error :%v\n", err)
			}

		}
	}()
}

func NotifAll(t time.Time) error {
	cert, err := certificate.FromP12File(config.GlobalConfiguration.AppleCertification, "")
	if err != nil {
		log.Fatalf("Cannot get certificate from Apple with error :%v\n", err)
	}
	fmt.Fprintf(os.Stdout, "Notify all apple device time :'%s'\n", t.Format("2006-01-02 15:04:05.000000"))
	devices, err := persistence.GetAllDevices()
	if err != nil {
		return err
	}
	var errorAppend string
	for _, d := range devices {
		n := &apns2.Notification{}
		n.DeviceToken = d.DeviceToken
		n.Topic = d.BundleId
		n.Payload = []byte(`{
			"aps" : {
				"content-available" : 1
			}
		}`)
		n.PushType = apns2.PushTypeBackground
		n.Priority = apns2.PriorityLow
		if strings.ToUpper(d.Type) == "ALERT" {
			n.Payload = []byte(d.Aps)
			n.PushType = apns2.PushTypeAlert
			n.Priority = apns2.PriorityHigh
			n.Expiration = time.Now().Add(2 * 60)
		}

		fmt.Fprintf(os.Stdout, "Sending notification on device token [%s]\n", d.DeviceToken)
		client := apns2.NewClient(cert).Development()
		res, err := client.Push(n)
		fmt.Fprintf(os.Stdout, "Response code [%d] : message body [%s]\n", res.StatusCode, res.Reason)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while sending apple notification on device token [%s] with error %v\n", d.DeviceToken, err)
			errorAppend += err.Error()
		} else {
			// save in database the date for this device token
			d.LastNotificationDate = time.Now()
			if err := persistence.UpdateDevice(d); err != nil {
				fmt.Fprintf(os.Stderr, "Error while updating date in database on device token [%s] with error %v\n", d.DeviceToken, err)
			}
		}
	}
	if errorAppend != "" {
		return fmt.Errorf(errorAppend)
	}
	return nil
}
