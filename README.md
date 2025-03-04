# Stay Alive

Stay Alive is a lightweight application designed to prevent your status in Microsoft Teams (or similar applications) from marking you as "away" due to inactivity. It accomplishes this by simulating periodic key presses in the background, keeping your status active.

## Features

- Automatically prevents your status from switching to "away."
- Customizable time interval between simulated key presses.
- Ability to select a specific F-Key for simulation.
- User-friendly settings accessible via a system tray icon.
- Configurations are saved automatically, so there's no need to reset the app each time.

---

## How to Use

1. **Download and Extract**  
   Simply download the application, extract the files, and run the executable file. There is no installation process required.

2. **Start the Application**  
   Run the binary file (executable) to launch the application. Upon startup:
    - A system tray icon for "Stay Alive" will appear.
    - The app will work silently in the background, emulating key presses based on the configured settings.

3. **Access the Settings**  
   Right-click or left-click on the "Stay Alive" icon in your system tray and select **Settings** to:
    - Modify the interval (in seconds) between key presses.
    - Choose the F-Key that should be simulated.
    - Save configurations.

4. **Save and Run**  
   After adjusting settings, press "Save," and the app will immediately apply the changes. These preferences are stored locally in the file `~/.stay_alive_config.json`.

---

## Configuration Options

1. **Key Press Interval**  
   You can customize the time interval between key presses (default: 60 seconds). This can be adjusted to any positive integer in seconds.

2. **Simulated Key**  
   The application allows you to select from a list of predefined F-Keys (`F13` to `F24`) to simulate during the operation.

3. **Configuration File**  
   All settings are saved to a JSON file at `~/.stay_alive_config.json`, which you can manually edit if needed.

Example of a configuration file:

```json
{
   "interval": 60,
   "key_code": 4221
}
```

---

## Requirements

- Tested on Windows operating systems.
- No additional tools are required to run this program.

---

## How to Stop the Application

To stop the application:
1. Locate the "Stay Alive" icon in the system tray.
2. Right-click on the icon and select **Exit**.

The app will terminate and no longer simulate key presses.

---

## Notes

- The simulated key press occurs in the background and does not affect your work or input.
- Frequent key simulation (e.g., a very low interval) might cause unexpected behavior in some programs. Adjust the interval accordingly.

Enjoy uninterrupted focus while staying "active"!

---