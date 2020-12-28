package notification

import (
	"fmt"
	"os"
	"time"

	"github.com/jeromelesaux/greenserver/config"
	"github.com/jeromelesaux/greenserver/persistence"
	"github.com/timehop/apns"
)

var (
	GlobalTicker *time.Ticker
	done         = make(chan bool)
)

func Initialise() {
	GlobalTicker = time.NewTicker(time.Hour)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-GlobalTicker.C:
				NotifAll(t)
			}
		}
	}()
}

func NotifAll(t time.Time) error {
	fmt.Fprintf(os.Stdout, "Notify all apple device time :'%s'\n", t.Format("2006-01-02 15:04:05.000000"))
	devices, err := persistence.GetAllDevices()
	if err != nil {
		return err
	}
	var errorAppend string
	for _, d := range devices {
		c, _ := apns.NewClient(apns.ProductionGateway,
			config.GlobalConfiguration.AppleCertification,
			config.GlobalConfiguration.AppleKey)

		p := apns.NewPayload()
		p.APS.Alert.Body = "Eco Data notification!"
		p.APS.Badge.Set(5)
		p.APS.Sound = "turn_down_for_what.aiff"

		m := apns.NewNotification()
		m.Payload = p
		m.DeviceToken = d.DeviceToken
		m.Priority = apns.PriorityImmediate
		fmt.Fprintf(os.Stdout, "Sending notification on device token [%s]\n", d.DeviceToken)
		if err := c.Send(m); err != nil {
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
