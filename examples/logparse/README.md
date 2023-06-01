## Logparse engine

A simple GPT bot designed to be installed on a server with logs that might need
to be analysed for errors. The bot's config contains a list of services and a
log file for each, which will be used by the bot for analysis.

https://github.com/venkytv/botmand/assets/718613/b6955559-4be8-4591-af9e-c3a0823bad17

See [logparse.yaml.tmpl](logparse.yaml.tmpl) for an example of the config for
the log files.

This bot is built using the  [langchain framework](https://python.langchain.com/en/latest/).

### Dependencies

- [Docker](https://www.docker.com/)
- [Python 3](https://www.python.org/)

### Installation

```
make install
```

This copies the bot's config as well as the bot script to the
default config directory: `~/botmand-engines`.

Edit `~/botmand-engines/logparse.yaml` and set the `OPENAI_API_KEY` variable.
Check [OpenAI documentation](https://openai.com/blog/openai-api) on signing up
for the API and [retrieving your API
key](https://help.openai.com/en/articles/4936850-where-do-i-find-my-secret-api-key).

Also configure the list of `LOG_xxx` parameters for the list of services to be
monitored.

If `botmand` is already running, make it reload its engines:

```
pkill -HUP botmand
```

### Usage

The bot only listens to messages directed at it in specific channels (by
default, #logparse). See `~/botmand-engines/logparse.yaml` after installation
to configure this.

To interact with it, ask it a question about one or more of the services
configured in the config file.  For instance, try "@Botman Does the webserver
look okay?"
