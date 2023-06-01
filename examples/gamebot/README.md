## Game bot

A simple bot which plays a choice of guessing games with the user.
The bot mainly serves to illustrate how to dynamically change behaviour
from threaded mode to channel mode and vice cersa.

Mode switches are handled by including one of the following commands
in a message:
- `botmand://switch/channel`: Switch to channel mode
- `botmand://switch/thread`: Create a new thread and switch to that

https://github.com/venkytv/botmand/assets/718613/ef5eb96f-5e41-48c9-bbf8-749f34981312

### Installation

```
make install
```

This copies the bot's config as well as the bot script to the
default config directory: `~/botmand-engines`.

If `botmand` is already running, make it reload its engines:

```
pkill -HUP botmand
```

### Usage

The bot is triggered in any channel where it is present by a directed message
with the word “game” in it.
