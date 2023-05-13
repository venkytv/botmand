# Botters Engine

Botters is a bot engine designed to siplify the process of creating Slack bots.
Botters bots handle a single conversation, reading input on stdin and writing
responses on stdout. (Think [inetd](https://en.wikipedia.org/wiki/Inetd) for
Slack conversations.) Botters launches multiple instances of the bots on demand,
one for each individual conversation.  Conversations terminate when the bot
exits.

Botters bots can be very simple. This is the complete code for a working bot
which keeps track of when was the last time someone mentioned "ChatGPT" in the
channel:

```bash
#!/bin/bash

LAST_MENTION=

while :; do
    read MSG
    if [[ "$MSG" = *ChatGPT* ]]; then
        NOW=$( date +'%s' )
        if [[ -n "$LAST_MENTION" ]]; then
            echo "It has been $(( NOW - LAST_MENTION )) seconds since someone mentioned ChatGPT last!"
        fi
        LAST_MENTION="$NOW"
    fi

    sleep 0.1
done
```

See [examples](examples) for more examples.
