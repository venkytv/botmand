## Mention bot engine

A simple channel bot that listens for a specific word to be mentioned in the
channel and responds with the time since the words was last mentioned.

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

The bot listens for any mention of the specified string ("ChatGPT" by default)
and prints a note about how long it has been since that string was last
mentioned in the channel.
