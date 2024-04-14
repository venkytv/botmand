## GPT chat bot engine

A simple GPT bot which, when asked directly, participates in the conversation
using previous messages in the conversation as context.  The bot is built
using the [OpenAI Python library](https://github.com/openai/openai-python).

**IMPORTANT: The bot sends the last few messages in the conversation (for up to
30 minutes in the past by default) to OpenAI to generate a response. Please
make sure you are aware of any privacy implications of that.**

Once invited to a channel, the bot keeps track of the current conversation.
When it encounters a message mentioning it, it responds using GPT to generate
the response.

Note that while the bot keeps track of the current conversation even if it is
not directly mentioned in the messages, it does **NOT** send anything to OpenAI
until it is specifically menioned in the conversation.

https://github.com/venkytv/botmand/assets/718613/acc6a5d6-1c6a-4c37-99a2-97a5e149ac6a

### Dependencies

- [Docker](https://www.docker.com/)
- [Python 3](https://www.python.org/)

### Installation

```
make install
```

This copies the bot's config as well as the bot script to the
default config directory: `~/botmand-engines`.

Edit `~/botmand-engines/gptbot.yaml` and set the `OPENAI_API_KEY` variable.
Check [OpenAI documentation](https://openai.com/blog/openai-api) on signing up
for the API and [retrieving your API
key](https://help.openai.com/en/articles/4936850-where-do-i-find-my-secret-api-key).

If `botmand` is already running, make it reload its engines:

```
pkill -HUP botmand
```

### Usage

At any point in any channel the bot is listening in (by default, #general),
send a message to the bot by @mentioning it and ask it any question. The bot
uses the last few messages in the conversation along with the question and uses
OpenAI to generate a response.
