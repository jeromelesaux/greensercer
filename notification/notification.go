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
	fmt.Fprintf(os.Stdout, "[NOTIFICATION] Initialise ticker each one hour.\n")
	go func() {
		for t := range GlobalTicker.C {
			if err := NotifAll(t); err != nil {
				fmt.Fprintf(os.Stderr, "[NOTIFICATION] Error while trying to notify devices with error :%v\n", err)
			}

		}
	}()
}

func PushBackgroundNotification(d *persistence.DeviceTable) error {
	cert, err := certificate.FromP12File(config.GlobalConfiguration.AppleCertification, "")
	if err != nil {
		log.Fatalf("[NOTIFICATION] Cannot get certificate from Apple with error :%v\n", err)
	}
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
	fmt.Fprintf(os.Stdout, "[NOTIFICATION]  Sending notification on device token [%s]\n", d.DeviceToken)
	client := apns2.NewClient(cert).Development()
	res, err := client.Push(n)
	if res != nil {
		fmt.Fprintf(os.Stdout, "[NOTIFICATION] Response code [%d] : message body [%s] and ApnsID [%s]\n", res.StatusCode, res.Reason, res.ApnsID)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "[NOTIFICATION] Error while sending apple notification on device token [%s] with error %v\n", d.DeviceToken, err)
		return err
	}
	return nil
}

func PushAlertNotification(d *persistence.DeviceTable, aps []byte) error {
	cert, err := certificate.FromP12File(config.GlobalConfiguration.AppleCertification, "")
	if err != nil {
		log.Fatalf("[NOTIFICATION] Cannot get certificate from Apple with error :%v\n", err)
	}
	n := &apns2.Notification{}
	n.DeviceToken = d.DeviceToken
	n.Topic = d.BundleId
	n.Payload = aps
	n.PushType = apns2.PushTypeAlert
	n.Priority = apns2.PriorityHigh
	n.Expiration = time.Now().Add(2 * 60)
	fmt.Fprintf(os.Stdout, "[NOTIFICATION]  Sending notification on device token [%s]\n", d.DeviceToken)
	client := apns2.NewClient(cert).Development()
	res, err := client.Push(n)
	if res != nil {
		fmt.Fprintf(os.Stdout, "[NOTIFICATION] Response code [%d] : message body [%s] and ApnsID [%s]\n", res.StatusCode, res.Reason, res.ApnsID)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "[NOTIFICATION] Error while sending apple notification on device token [%s] with error %v\n", d.DeviceToken, err)
		return err
	}
	return nil
}

func NotifAll(t time.Time) error {

	fmt.Fprintf(os.Stdout, "[NOTIFICATION] Notify all apple device time :'%s'\n", t.Format("2006-01-02 15:04:05.000000"))
	devices, err := persistence.GetAllDevices()
	if err != nil {
		return err
	}
	var errorAppend string
	for _, d := range devices {
		if strings.ToUpper(d.Type) == "ALERT" {
			aps := []byte(`{"aps":{
					"alert":{"title":"Push Notification","subtitle":"Test Push Notifications","body":"Testing Push Notifications on iOS Simulator"}
					}}`)
			if err := PushAlertNotification(d, aps); err != nil {
				errorAppend += err.Error()
			} else {
				// save in database the date for this device token
				d.LastNotificationDate = time.Now()
				if err := persistence.UpdateDevice(d); err != nil {
					fmt.Fprintf(os.Stderr, "[NOTIFICATION] Error while updating date in database on device token [%s] with error %v\n", d.DeviceToken, err)
				}
			}
		}
	}
	if errorAppend != "" {
		return fmt.Errorf(errorAppend)
	}
	return nil
}

func Notify(deviceToken string, aps []byte) error {

	fmt.Fprintf(os.Stdout, "[NOTIFICATION] Notify the apple device [%s] time :'%s'\n", deviceToken, time.Now().Format("2006-01-02 15:04:05.000000"))
	d, err := persistence.GetDeviceByToken(deviceToken)
	if err != nil {
		return err
	}
	err = PushAlertNotification(d, aps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[NOTIFICATION] Error while sending apple notification on device token [%s] with error %v\n", d.DeviceToken, err)
		return err
	} else {
		// save in database the date for this device token
		d.LastNotificationDate = time.Now()
		if err := persistence.UpdateDevice(d); err != nil {
			fmt.Fprintf(os.Stderr, "[NOTIFICATION] Error while updating date in database on device token [%s] with error %v\n", d.DeviceToken, err)
		}
	}
	return nil
}
