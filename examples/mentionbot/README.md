## Basic bot engine

A simple channel bot that listens for a specific word to be mentioned in the
channel and responds with the time since the words was last mentioned.

### Installation

```
make install
```

This copies the bot's config as well as the bot script to the
default config directory: `~/teabot-engines`.

If `teabot` is already running, make it reload its engines:

```
pkill -HUP teabot
```
