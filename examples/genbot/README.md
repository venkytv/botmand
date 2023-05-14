## Genbot engine

A simple GPT bot which can generate a shell script to perform the required
task. The bot then replaces itself with the generated shell script.

The bot only listens to messages directed at it in specific channels
(by default, #genbot). See `~/teabot-engines/genbot.yaml` after installation
to configure this.

Currently running genbots can be listed by the command "genbot ls", which
prints a list of running genbots along with their IDs.
Any running genbot can be stopped by entering the command: "genbot stop <ID>".

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
