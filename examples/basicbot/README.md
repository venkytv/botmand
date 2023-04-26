## Basic bot engine

A very basic bot engine which sets up threaded conversations
where it echoes the user's text back to them.

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
