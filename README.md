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

Botters is designed primarily for supporting conversational bots, which
incidentally happens to be how most current AI chatbots are set up. For
examples of how you can use Botters to develop AI-powered chatbots, have a look
at the [gptbot](examples/gptbot) and [genbot](examples/genbot) examples.

## Installation

* Download the [latest release](../../releases/latest) of botters for your platform.
* Extract the archive and move the `botters` executable to a directory of your choice.

## Quickstart

### Generate a Slack classic app bot token

1. Create a Classic Slack app: https://api.slack.com/apps?new_classic_app=1
2. Click on "App Home" in the left sidebar to get to the bot creation page.
3. Click on the "Add Legacy Bot User" button.
4. Set a display name and user name for the bot and add the bot.
5. Click on "Install App" in the left sidebar.
6. Click "Install to Workspace" and allow the bot access to the workspace.
7. At this point, you should see the "OAuth Tokens" for your workspace.
    There are two tokens displayed, but you just need the "Bot User OAuth Token" which
    should start with the string `xoxb-`.
8. Copy this token to the file `.slack.token` in your home directory.

### Set up the example basicbot

```bash
git clone https://github.com/venkytv/botters.git
cd botters/examples/basicbot
make install
```

### Invite the Slack bot to a channel in your workspace

Go to a channel in Slack and either use the `/invite` command, or just message
the new Slack bot to invite it to the channel.

### Start botters

Launch `botters` from the directory you moved it to.

```bash
./botters
```

### Start a conversation

Send a "hello" message mentioning the bot name. The bot should respond in a
thread echoing what you said. And then continue echoing whatever you say in the
thread until you say "bye".

### Explore more real-world bots

* [gptbot](examples/gptbot): An AI chat assistant which can participate in the
  conversation when asked to.
* [genbot](examples/genbot): An AI assistant which can generate shell scripts
  for requested tasks and run them within the bot.

### Write your own bot

The [bot writing guide](BOT-WRITING-GUIDE.md) has details on writing bots.  You
could also have a look at the included [example bots](examples).
