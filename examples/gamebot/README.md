## Game bot

A simple bot which plays a choice of guessing games with the user.
The bot mainly serves to illustrate how to dynamically change behaviour
from threaded mode to channel mode and vice cersa.

Mode switches are handled by including one of the following commands
in a message:
- `botters://switch/channel`: Switch to channel mode
- `botters://switch/thread`: Create a new thread and switch to that

### Installation

```
make install
```

This copies the bot's config as well as the bot script to the
default config directory: `~/botters-engines`.

If `botters` is already running, make it reload its engines:

```
pkill -HUP botters
```

### Usage

The bot is triggered in any channel it is in by a directed message which
the word "game" in it.
