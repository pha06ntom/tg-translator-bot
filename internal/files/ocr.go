package files

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runTesseract(imgPath string) (string, error) {
	cmd := exec.Command(
		"tesseract",
		imgPath,
		"stdout",
		"-l", "rus+eng",
		"--oem", "1",
		"--psm", "11",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tesseract: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func convertTIFFToPNG(path string) (string, error) {
	tmpFile, err := os.CreateTemp("", "tiff_ocr_*.png")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()

	cmd := exec.Command(
		"convert",
		path,
		"-auto-orient",
		"-colorspace", "Gray",
		"-density", "300",
		"-resize", "200%",
		"-deskew", "40%",
		"-normalize",
		"-contrast-stretch", "0",
		"-sharpen", "0x1.0",
		"-threshold", "60%",
		"-strip",
		tmpPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("convert tiff to png: %v: %s", err, strings.TrimSpace(string(out)))
	}

	return tmpPath, nil
}

func cropBottomArea(path string) (string, error) {
	tmpFile, err := os.CreateTemp("", "ocr_bottom_*.png")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()

	cmd := exec.Command(
		"convert",
		path,
		"-gravity", "south",
		"-crop", "100%x40%+0+0",
		"+repage",
		"-colorspace", "Gray",
		"-resize", "200%",
		"-normalize",
		"-contrast-stretch", "0",
		"-sharpen", "0x1.0",
		"-threshold", "60%",
		"-strip",
		tmpPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("crop bottom area: %v: %s", err, strings.TrimSpace(string(out)))
	}

	return tmpPath, nil
}
