## Basic bot engine

A simple GPT bot which, when asked directly, participates in the conversation
using previous messages in the conversation as context.

**IMPORTANT: The bot sends the last few messages in the conversation (for up to
30 minutes in the past by default) to OpenAI to generate a response. Please
make sure you are aware of any privacy implications of that.**

Once invited to a channel, the bot keeps track of the current conversation.
When it encounters a message mentioning it, it responds using GPT to generate
the response.

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
