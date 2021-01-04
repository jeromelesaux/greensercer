package persistence

import "time"

type DeviceTable struct {
	Uid                  string
	DeviceToken          string
	BundleId             string
	Type                 string
	Aps                  string
	LastNotificationDate time.Time
}

func NewDeviceTable(token, bundleid, notificationType, aps string) *DeviceTable {
	return &DeviceTable{
		DeviceToken: token,
		BundleId:    bundleid,
		Type:        notificationType,
		Aps:         aps,
	}
}

func DeviceTableRaw(uid, deviceToken, bundleid, notificationType, aps string, lastDate time.Time) *DeviceTable {
	return &DeviceTable{
		Uid:                  uid,
		DeviceToken:          deviceToken,
		BundleId:             bundleid,
		Type:                 notificationType,
		Aps:                  aps,
		LastNotificationDate: lastDate,
	}
}
