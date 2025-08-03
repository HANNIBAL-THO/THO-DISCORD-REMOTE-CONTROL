package utils

import (
	"fmt"
	"image/png"
	"os"
	"time"

	"github.com/kbinani/screenshot"
)

func CaptureScreen() (string, error) {
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("screenshot_%d.png", time.Now().Unix())
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return "", err
	}

	return filename, nil
}
