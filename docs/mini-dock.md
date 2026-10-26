# Mini dock

Enable **Mini dock when minimized** in **Settings → App** to keep a small session overview over other apps when minimizing enowX. It is off by default. Closing the main window still follows the existing tray preference; closing it does not show the dock.

The compact dock counts working sessions and sessions needing attention (waiting for input or exited with an error). Click the summary to expand its list, ordered by attention, working, then the rest. **Host** filters the list by This PC or a saved host; it does not change the summary's totals. Long lists scroll. A session's status can update while the pointer or keyboard focus is over the list, but its row stays in place until the user leaves it.

Click a session to restore the main window on that session, connecting to its host first when needed. The restore icon and **Open main window** bring back the current workspace. Escape collapses the list. Restoring enowX from its taskbar or tray also hides the dock. Showing the dock does not take focus from another app.

Drag its heading or grip to move it. Its left and bottom edges are remembered on this device, and brought back onto a display if the monitor arrangement changes. Expanding preserves the bottom edge, opening upwards within the display's work area. Disable the preference to hide the dock and return to normal minimize behavior.

The dock fades and grows into place, animates expansion, and briefly highlights attention when a session starts needing input or fails. **Reduce mini dock motion**, below the enabling switch, removes those animations. Both preferences persist locally between launches.

This view uses the app's existing daemon snapshots and background host watches, including their normal refresh interval. No extra connection, RPC, or network polling is introduced. Unreachable hosts keep their last known sessions marked **Offline**, without counting them as currently working or needing attention. Main sessions are listed once; their shell tabs and the agents embedded in Notes and Canvas are not separate dock entries.
