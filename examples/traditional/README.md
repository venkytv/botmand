## Traditional bot engine

A traditional bot which only responds to direct messages. These are simple to
write with most bot frameworks and do not play to botters's strengths. This is
just a demo to show that traditional bot interactions are also supported.

This bots listens for triggers matching the phrase which includes the words
"what" and "time" and replies with the local time.

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

Once installed, go to a channel the bot is installed in and type a directed
message at the bot saying "what is the time?".
