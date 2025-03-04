package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"golang.org/x/sys/windows/registry"
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
	Interval  int  `json:"interval"`
	KeyCode   int  `json:"key_code"`
	AutoStart bool `json:"auto_start"`
}

const (
	appName               = "StayAlive"
	defaultKeyCode        = keybd_event.VK_F15
	defaultConfigFileName = ".stay_alive_config.json"
)

var (
	configMu        sync.Mutex
	intervalMu      sync.Mutex
	keyLoopCancel   context.CancelFunc
	keyLoopCtx      context.Context
	currentInterval = 60 * time.Second
	currentKey      = defaultKeyCode
	keyMapping      = map[string]int{
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
	reverseKeyMapping = func() map[int]string {
		mapping := make(map[int]string)
		for key, code := range keyMapping {
			mapping[code] = key
		}
		return mapping
	}()
)

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

	go reloadService()
	app.Main()
}

func getUserHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error retrieving the home directory:", err)
	}
	return homeDir
}

func getConfigPath() string {
	return filepath.Join(getUserHomeDirectory(), defaultConfigFileName)
}

func saveConfig() {
	configMu.Lock()
	defer configMu.Unlock()
	config := Config{
		Interval:  int(currentInterval.Seconds()),
		KeyCode:   currentKey,
		AutoStart: isAutoStartEnabled(),
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

	go reloadService()
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
		config.KeyCode = defaultKeyCode
	}

	currentInterval = time.Duration(config.Interval) * time.Second
	currentKey = config.KeyCode

	if config.AutoStart && !isAutoStartEnabled() {
		if enableAutoStart() != nil {
			log.Println("Error enabling AutoStart:", err)
		}
	} else if !config.AutoStart && isAutoStartEnabled() {
		if disableAutoStart() != nil {
			log.Println("Error disabling AutoStart:", err)
		}
	}
}

func keyLoop(ctx context.Context, kb keybd_event.KeyBonding) {
	for {
		select {
		case <-ctx.Done(): // Überprüfen, ob der Kontext abgebrochen wurde
			log.Println("Key loop stopped.")
			return
		default:
			err := kb.Launching()
			if err != nil {
				log.Println("Key press error:", err)
			} else {
				log.Println("Key pressed:", currentKey)
			}

			intervalMu.Lock()
			interval := currentInterval
			intervalMu.Unlock()

			time.Sleep(interval)
		}
	}
}

func reloadService() {
	loadConfig()

	if keyLoopCancel != nil {
		log.Println("Cancelling current key loop...")
		keyLoopCancel()
	}

	keyLoopCtx, keyLoopCancel = context.WithCancel(context.Background())
	keyboard, err := keybd_event.NewKeyBonding()
	if err != nil {
		log.Fatal("Keyboard error:", err)
	}
	keyboard.SetKeys(currentKey)
	go keyLoop(keyLoopCtx, keyboard)

	log.Println("Service reloaded")
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
			app.Size(unit.Dp(400), unit.Dp(200)),
		)

		var (
			ops       op.Ops
			theme     = material.NewTheme()
			interval  widget.Editor
			keySelect widget.Enum
			btn       widget.Clickable
			autoStart widget.Bool
			dropDown  widget.Clickable
			errText   string
		)

		intervalMu.Lock()
		interval.SetText(fmt.Sprint(int(currentInterval.Seconds())))
		autoStart.Value = isAutoStartEnabled()
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

						if autoStart.Value {
							if err := enableAutoStart(); err != nil {
								errText = fmt.Sprintf("Error enabling AutoStart: %s", err)
							}
						} else {
							if err := disableAutoStart(); err != nil {
								errText = fmt.Sprintf("Error disabling AutoStart: %s", err)
							}
						}

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
							return material.CheckBox(theme, &autoStart, "Enable AutoStart").Layout(gtx)
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

func enableAutoStart() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exePath = strings.ReplaceAll(exePath, `"`, `""`)

	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.WRITE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer func(key registry.Key) {
		err := key.Close()
		if err != nil {
			log.Println("Error closing registry key:", err)
		}
	}(key)

	err = key.SetStringValue(appName, exePath)
	if err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	fmt.Println("Autostart successfully enabled", exePath)
	return nil
}

func disableAutoStart() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.WRITE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer func(key registry.Key) {
		err := key.Close()
		if err != nil {
			log.Println("Error closing registry key:", err)
		}
	}(key)

	err = key.DeleteValue(appName)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			fmt.Println("Autostart already disabled (no entry found).")
			return nil
		}
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	fmt.Println("Autostart successfully disabled.")
	return nil
}

func isAutoStartEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.READ)
	if err != nil {
		log.Println("Error opening registry key:", err)
		return false
	}
	defer func(key registry.Key) {
		err := key.Close()
		if err != nil {
			log.Println("Error closing registry key:", err)
		}
	}(key)

	_, _, err = key.GetStringValue(appName)
	if err != nil {
		return false
	}

	return true
}
