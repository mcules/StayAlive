# Stay Alive

Stay Alive is a lightweight desktop application designed to prevent your presence status in Microsoft Teams (or similar apps) from switching to "Away" due to inactivity. It achieves this by simulating periodic key presses in the background.

## Features

- **Prevent away status automatically:** Simulates key presses in the background to keep your status active.
- **Customizable time interval:** Users can set the interval (in seconds) between key presses.
- **Customizable key selection:** Allows choice of function keys (F13 to F24) for simulation.
- **AutoStart support:** Option to launch the app automatically on system startup.
- **System Tray integration:** A tray icon provides quick access to settings and app control.
- **GUI for configuration:** A user-friendly interface to modify the interval, select keys, and enable/disable AutoStart.
- **Config file storage:** Stores all settings in `~/.stay_alive_config.json`.

---

## Requirements

- **Operating System:** Windows (AutoStart functionality uses Windows Registry).
- **No installation necessary:** The app is portable and requires no additional dependencies.

---

## Installation and Usage

1. **Launch the application**  
   Download and extract the executable file, then run it. After startup:
   - A tray icon appears in the system tray to indicate that "Stay Alive" is active.
   - The application runs silently in the background.

2. **Open the configuration settings**  
   Right-click the tray icon and select **Settings** to:
   - Adjust the interval (in seconds) for key presses.
   - Choose a key from F13 to F24.
   - Enable or disable the AutoStart option.

3. **Save configuration**  
   After making changes, click **Save**. The new settings apply immediately and are stored automatically in the `~/.stay_alive_config.json` file.

4. **Enable AutoStart**  
   When the **AutoStart** option is enabled, the app will automatically start on the next system boot.

---

## Technical Details

### Configuration File:
- **Path:**
  ```plaintext
  %USERPROFILE%\.stay_alive_config.json
  ```
  This file stores the interval, selected key, and AutoStart preferences in JSON format.

### Example Configuration File:
```json
{
  "interval": 120,
  "key_code": 114,
  "auto_start": true
}
```

### Supported F-Keys:
The app supports F-keys ranging from F13 to F24 for simulation. Key selection works via a mapping in the code.

---

## How to Use

1. **Start the app:**  
   Double-click the executable file to start the application. A tray icon will appear in your system tray.

2. **Modify settings:**
   - Right-click the tray icon → Select **Settings**.
   - In the GUI window, configure the interval, select a key, and enable/disable AutoStart.

3. **Stop the app:**  
   Right-click the tray icon and choose **Exit** to terminate the application.

---

## Notes

- Changes made in the configuration take effect immediately without requiring a restart.
- The automatic key simulation runs in the background and does not interfere with normal usage of your computer.
- The GUI settings editor includes a dropdown for key selection and provides warnings for invalid inputs (e.g., non-numeric intervals).

---

## Changelog

### What's new in the current version:
- **AutoStart Support:** Added the capability to launch the app automatically on each system startup.
- **GUI Settings:** A fully interactive graphical interface for live configuration adjustments.
- **Improved Key Selection:** Includes a dropdown for an easier selection of function keys.

---

If you need additional adjustments or further explanations, feel free to ask! 😊