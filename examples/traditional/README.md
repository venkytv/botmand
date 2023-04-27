## Traditional bot engine

A traditional bot which only responds to direct messages. These are simple to
write with most bot frameworks and do not play to teabot's strengths. This is
just a demo to show that traditional bot interactions are also supported.

This bots listens for triggers matching the phrase which includes the words
"what" and "time" and replies with the local time.

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
