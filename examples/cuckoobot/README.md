## Cuckoo bot engine

A simple bot which prints the message "cuckoo" on the hour every hour in any
channel it is invited into.

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

The bot needs no interaction. It is activated in every channel it is present in
by any messages there, and once activated, it will print out a "cuckoo" on the
hour every hour.
