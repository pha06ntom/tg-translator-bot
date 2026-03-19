package telegram

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) downloadTelegramFile(fileID, dstPath string) error {
	f, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return err
	}

	url, err := b.api.GetFileDirectURL(f.FileID)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram file download failed: %s", resp.Status)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func sanitizeName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.Trim(name, "._-")
	if name == "" {
		return "file"
	}
	return name
}
