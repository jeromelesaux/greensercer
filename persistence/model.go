package persistence

import "time"

type DeviceTable struct {
	Uid                  string
	DeviceToken          string
	LastNotificationDate time.Time
}

func NewDeviceTable(token string) *DeviceTable {
	return &DeviceTable{DeviceToken: token}
}

func DeviceTableRaw(uid string, deviceToken string, lastDate time.Time) *DeviceTable {
	return &DeviceTable{
		Uid:                  uid,
		DeviceToken:          deviceToken,
		LastNotificationDate: lastDate,
	}
}
