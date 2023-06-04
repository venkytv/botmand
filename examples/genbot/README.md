## Genbot engine

A simple GPT bot which can generate a shell script to perform the required
task. The generated script is run within a docker container, with any files
or directories it requires from the host mounted as read-only volumes.

https://github.com/venkytv/botmand/assets/718613/a0553be7-9c7b-4bc5-892a-d19b24c12f8d

### Dependencies

- [Docker](https://www.docker.com/)
- [Python 3](https://www.python.org/)

### Installation

```
make install
```

This copies the bot's config as well as the bot script to the
default config directory: `~/botmand-engines`.

Edit `~/botmand-engines/genbot.yaml` and set the `OPENAI_API_KEY` variable.
Check [OpenAI documentation](https://openai.com/blog/openai-api) on signing up
for the API and [retrieving your API
key](https://help.openai.com/en/articles/4936850-where-do-i-find-my-secret-api-key).

If `botmand` is already running, make it reload its engines:

```
pkill -HUP botmand
```

### Usage

The bot only listens to messages directed at it in specific channels
(by default, #genbot). See `~/botmand-engines/genbot.yaml` after installation
to configure this.

When issued an instruction to perform a specific task (eg., to monitor a
specific file and print an alert when its contents change), the bot generates a
shell script which it then runs by the user for confirmation. Once the user
accepts the script, it generates a genbot (a docker image with the script) and
runs it.

Currently running genbots can be listed by the command "genbot ls", which
prints a list of running genbots along with their IDs. Any running genbot can
be stopped by entering the command: "genbot stop <ID>".
