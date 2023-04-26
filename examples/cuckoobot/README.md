## Basic bot engine

A simple bot which prints the message "cuckoo" on the hour every hour in any
channel it is invited into.

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
