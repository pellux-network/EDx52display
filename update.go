package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ncruces/zenity"
	log "github.com/sirupsen/logrus"
)

const githubRepo = "pellux-network/EDx52display"

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func CheckForUpdate(currentVersion string) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	resp, err := client.Get(url)
	if err != nil {
		log.Warnf("Update check failed: %v", err)
		return
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Warnf("Failed to parse update info: %v", err)
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")
	if latest != "" && latest != current {
		log.Infof("A new version is available: %s (current: %s). Download at: %s", release.TagName, currentVersion, release.HTMLURL)
		showUpdatePopup(release.TagName, release.HTMLURL, currentVersion)
	} else {
		log.Infof("You are running the latest version: %s", currentVersion)
	}
}

func showUpdatePopup(version, url, currentVersion string) {
	msg := fmt.Sprintf(
		"A new version (%s) is available! You are on version %s\n\nRelease page:\n%s",
		version, currentVersion, url,
	)
	title := "EDx52Display Update Available"

	err := zenity.Question(
		msg,
		zenity.Title(title),
		zenity.OKLabel("Update"),
		zenity.ExtraButton("Release Page"),
		zenity.CancelLabel("Dismiss"),
	)
	switch err {
	case zenity.ErrExtraButton:
		openBrowser(url)
	case nil:
		// TODO: Implement auto-update logic here
		// For now, just open the release page
		openBrowser(url)
	case zenity.ErrCanceled:
		// Dismiss (user clicked X or Cancel)
	}
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // linux, freebsd, etc.
		cmd = "xdg-open"
		args = []string{url}
	}
	_ = exec.Command(cmd, args...).Start()
}
