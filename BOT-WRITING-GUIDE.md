# Guide to writing bots

BotManD bots are simple executables which read input on stdin and respond on
stdout. BotManD bots handle a single conversation. The task of managing separate
conversations is handled by BotManD. This makes writing bots a trivial task.

The following is a bot which waits for mentions of the word "ChatGPT" and chips
in with a comment on how long it has been since someone mentioned it last.

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

The bot can be located anywhere as long as it is executable by the user BotManD
is running as.  BotManD loads bots using config files located (by default) in
`~/botmand-engines`.

Each bot needs a config file which configures the bot's behaviour. The config
file for the bot above could be:

```yaml
handler: /home/foo/bin/mentionbot.sh  # Required; path to bot executable
direct-message-triggers-only: false   # Do not wait to be specifically mentioned
                                      # to start listening in
```

See the [sample config file](examples/sample-config.yaml) for details on all
available options and defaults. The config files for the [examples](examples)
also illustrate most common config options.

## Environment variables

The bot will have access to the following environment variables:

* `BOTTERS_USER_ID`: ID of the bot user account; normally used to identify messages which mention the bot
* `BOTTERS_USER_NAME` User name of the bot user account
* `BOTTERS_CHANNEL`: Name of the channel this bot instance is running in
* `BOTTERS_CHANNEL_ID`: ID of the channel this bot instance is running in
* `BOTTERS_LOCALE`: Locale of the channel the bot is running in

See [gptbot](examples/gptbot/gptbot.py) for an example of how a bot might use these variables.

## Special Handling

BotManD primarily serves as a conduit for relaying user messages to the bot and
delivering bot messages to the user. However, there are a few bot messages that
are handled differently.

The first is a message from the bot with the text `...`. This message is not
posted on the channel, but instead a typing indicator is generated. This is
useful in cases where the bot might take some time to generate a response, and
it wants to indicate to the user that it is doing something.

The second set of special messages are BotManD commands.

### BotManD commands

BotManD uses a `botmand://` URL scheme for commands. A bot can include these
commands in messages to instruct BotManD to perform specific actions.  The
following commands are supported currently:

* `botmand://switch/channel`: Switch bot from threaded mode to channel mode.
  After this command is received, BotManD abandons the thread the bot is
  interacting in and switches further interaction to the channel.
* `botmand://switch/thread`: Switch bot from channel mode to threaded mode.
  When this command is received in a message, BotManD creates a new thread from
  the message text (excluding the command). For instance, the following response
  from the bot "Creating new thread botmand://switch/thread" will post the
  message "Creating new thread" to Slack, and switch to threaded mode using that
  message. All subsequent messages by the bot are delivered on that thread.
  * Note that once the mode is switched to "thread", the bot does not listen on
    the channel any more. But until another message is sent by the bot (or by a
    user explicitly creating a Slack thread from that message), the actual
    thread is not seen in Slack. This can be confusing. The best thing to do
    after switching to "thread" mode is for the bot to send a follow-up
    message. The follow-up message will automatically create the thread in
    Slack.

For an example on using BotManD commands, see [gamebot](examples/gamebot).

## Things to keep in mind

* Make sure the bot executable is either line-buffered or unbuffered.
  Fully buffered output might mean that the bot's output might not be delivered
  to BotManD until a block if filled. Check the [GNU Buffering Concepts
  manual](https://www.gnu.org/software/libc/manual/html_node/Buffering-Concepts.html)
  for more details.
  * Shell script are line-buffered by default, but for bots written in other
    languages, you might need to either switch to unbuffered mode or perform
    explicit flushes while writing to stdout.  For example, you can run a python
    script with `#!/usr/bin/python -u` for unbuffered mode, or explicitly flush
    each print with something like `print("my message", flush=True)`.
