package push

import (
	log "github.com/Sirupsen/logrus"
	"github.com/timehop/apns"
)

// APNSPusher pushes notification via apns
type APNSPusher struct {
	// we are directly coupling on apns as it seems redundant to duplicate
	// all the payload and client logic and interfaces.
	Client apns.Client
}

// NewAPNSPusher returns a new APNSPusher for use
func NewAPNSPusher(gateway string, certPath string, keyPath string) (*APNSPusher, error) {
	client, err := apns.NewClientWithFiles(gateway, certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &APNSPusher{Client: client}, nil
}

// Init set up the notification error channel
func (pusher *APNSPusher) Init() error {
	go func() {
		for result := range pusher.Client.FailedNotifs {
			log.Errorf("Failed to send notification = %s: %v", result.Notif.ID, result.Err)
		}
	}()

	return nil
}

// Send sends a notification to the device identified by the
// specified deviceToken
func (pusher *APNSPusher) Send(m Mapper, deviceToken string) error {
	payload := apns.NewPayload()
	payload.APS.ContentAvailable = 1

	customMap := m.Map()
	for key, value := range customMap {
		if err := payload.SetCustomValue(key, value); err != nil {
			log.Errorf("Failed to set key = %v, value = %v", key, value)
		}
	}

	notification := apns.NewNotification()
	notification.Payload = payload
	notification.DeviceToken = deviceToken
	notification.Priority = apns.PriorityImmediate

	if err := pusher.Client.Send(notification); err != nil {
		log.Printf("Failed to send Push Notification: %v", err)
		return err
	}

	return nil
}
