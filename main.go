package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gioui.org/app"
	_ "gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	_ "gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/getlantern/systray"
	"github.com/micmonay/keybd_event"
)

type Config struct {
	Interval int `json:"interval"`
	KeyCode  int `json:"key_code"`
}

var (
	configMu        sync.Mutex
	intervalMu      sync.Mutex
	currentInterval = 60 * time.Second
	currentKey      = keybd_event.VK_F15
)

var keyMapping = map[string]int{
	"F13": keybd_event.VK_F13,
	"F14": keybd_event.VK_F14,
	"F15": keybd_event.VK_F15,
	"F16": keybd_event.VK_F16,
	"F17": keybd_event.VK_F17,
	"F18": keybd_event.VK_F18,
	"F19": keybd_event.VK_F19,
	"F20": keybd_event.VK_F20,
	"F21": keybd_event.VK_F21,
	"F22": keybd_event.VK_F22,
	"F23": keybd_event.VK_F23,
	"F24": keybd_event.VK_F24,
}
var reverseKeyMapping map[int]string

func init() {
	reverseKeyMapping = make(map[int]string)
	for key, code := range keyMapping {
		reverseKeyMapping[code] = key
	}
}

func main() {
	loadConfig()

	go func() {
		systray.Run(onReady, nil)
	}()

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		log.Fatal("Keyboard error:", err)
	}
	kb.SetKeys(currentKey)

	go keyLoop(kb)
	app.Main()
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Fehler beim Abrufen des Home-Verzeichnisses:", err)
	}
	return filepath.Join(homeDir, ".stay_alive_config.json")
}

func saveConfig() {
	configMu.Lock()
	defer configMu.Unlock()

	config := Config{
		Interval: int(currentInterval.Seconds()),
		KeyCode:  currentKey,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Println("Error saving configuration:", err)
		return
	}

	err = os.WriteFile(getConfigPath(), data, 0644)
	if err != nil {
		log.Println("Error writing configuration file:", err)
	}

	log.Println("Configuration saved to:", getConfigPath())
}

func loadConfig() {
	configMu.Lock()
	defer configMu.Unlock()

	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("No saved configuration found, using default values.")
			return
		}
		log.Println("Error reading configuration file:", err)
		return
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Println("Error parsing configuration file:", err)
		return
	}

	if config.Interval == 0 {
		config.Interval = 60
	}
	if config.KeyCode == 0 {
		config.KeyCode = keybd_event.VK_F15
	}

	currentInterval = time.Duration(config.Interval) * time.Second
	currentKey = config.KeyCode
}

func keyLoop(kb keybd_event.KeyBonding) {
	for {
		err := kb.Launching()
		if err != nil {
			log.Println("Key press error:", err)
		}

		intervalMu.Lock()
		interval := currentInterval
		intervalMu.Unlock()

		time.Sleep(interval)
	}
}

func onReady() {
	systray.SetIcon(getIconData())
	systray.SetTitle("Stay Alive")
	systray.SetTooltip("Stay Alive")

	mSettings := systray.AddMenuItem("Settings", "Open configuration")
	mQuit := systray.AddMenuItem("Exit", "Quit application")

	go func() {
		for {
			select {
			case <-mSettings.ClickedCh:
				openSettings()
			case <-mQuit.ClickedCh:
				systray.Quit()
				os.Exit(0)
			}
		}
	}()
}

func openSettings() {
	go func() {
		window := new(app.Window)

		window.Option(
			app.Title("Stay Alive Settings"),
			app.Size(unit.Dp(400), unit.Dp(120)),
		)

		var (
			ops       op.Ops
			theme     = material.NewTheme()
			interval  widget.Editor
			keySelect widget.Enum
			btn       widget.Clickable
			dropDown  widget.Clickable
			errText   string
		)

		intervalMu.Lock()
		interval.SetText(fmt.Sprint(int(currentInterval.Seconds())))
		intervalMu.Unlock()

		keys := []string{"F13", "F14", "F15", "F16", "F17", "F18", "F19", "F20", "F21", "F22", "F23", "F24"}
		keySelect.Value, _ = getKeyString(currentKey)

		for {
			switch e := window.Event().(type) {
			case app.DestroyEvent:
				return
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				if btn.Clicked(gtx) {
					intervalText := interval.Text()
					newInterval, err := strconv.Atoi(intervalText)
					if err != nil || newInterval <= 0 {
						errText = "Invalid interval"
					} else {
						intervalMu.Lock()
						currentInterval = time.Duration(newInterval) * time.Second
						currentKey, _ = getKeyCode(keySelect.Value)
						intervalMu.Unlock()
						errText = ""
					}

					saveConfig()
				}

				layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return material.Body1(theme, "Interval (seconds):").Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Stack{}.Layout(gtx,
										layout.Expanded(func(gtx layout.Context) layout.Dimensions {
											paint.FillShape(gtx.Ops, color.NRGBA{R: 245, G: 245, B: 245, A: 255},
												clip.Rect{Max: gtx.Constraints.Min}.Op(),
											)
											shadow := widget.Border{
												Color: color.NRGBA{R: 180, G: 180, B: 180, A: 100},
												Width: unit.Dp(2),
											}
											return shadow.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
												return material.Editor(theme, &interval, "").Layout(gtx)
											})
										}),
									)
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return material.Body1(theme, "Select F-Key:").Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if dropDown.Clicked(gtx) {
										for i, key := range keys {
											if key == keySelect.Value {
												nextIndex := (i + 1) % len(keys)
												keySelect.Value = keys[nextIndex]
												break
											}
										}
									}
									return material.Button(theme, &dropDown, strings.ToUpper(keySelect.Value)).Layout(gtx)
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Button(theme, &btn, "Save").Layout(gtx)
						}),

						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if errText != "" {
								return material.Body1(theme, errText).Layout(gtx)
							}
							return layout.Dimensions{}
						}),
					)
				})

				e.Frame(gtx.Ops)
			}
		}
	}()
}

func getKeyCode(key string) (int, error) {
	if code, ok := keyMapping[key]; ok {
		return code, nil
	}
	return 0, errors.New("unknown Key")
}

func getKeyString(code int) (string, error) {
	if key, ok := reverseKeyMapping[code]; ok {
		return key, nil
	}
	return "", errors.New("unknown Keycode")
}

func getIconData() []byte {
	data, err := os.ReadFile("favicon.ico")
	if err != nil {
		log.Fatal("Icon error:", err)
	}
	return data
}
