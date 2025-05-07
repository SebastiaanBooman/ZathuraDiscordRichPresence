package discordrpc

import (
	"github.com/hugolgst/rich-go/client"
	"time"
)

const (
	APP_ID = "1369674663914897453"
)

func Login() error {
	err := client.Login(APP_ID)
	if err != nil {
		return err
	}
	return nil
}

func Logout() error {
	client.Logout()
	return nil
}

func SetActivity(state string, details string, largeImage string, largeText string, startTime time.Time) error {
	err := client.SetActivity(client.Activity{
		State:      state,
		Details:    details,
		LargeImage: largeImage,
		LargeText:  largeText,
		SmallImage: "",
		SmallText:  "",
		Timestamps: &client.Timestamps{
			Start: &startTime,
		},
	})

	if err != nil {
		return err
	}

	return nil
}
