package event

import (
	"encoding/base64"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

func GenerateCheckInQRCode(
	event *Event,
	frontendURL string,
) (string, error) {

	url := fmt.Sprintf(
		"%s/check-in/%d/%s",
		frontendURL,
		event.ID,
		event.CheckInToken,
	)

	png, err := qrcode.Encode(
		url,
		qrcode.Medium,
		512,
	)
	if err != nil {
		return "", err
	}

	return "data:image/png;base64," +
		base64.StdEncoding.EncodeToString(png), nil
}
