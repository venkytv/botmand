## Basic bot engine

A very basic bot engine which sets up threaded conversations
where it echoes the user's text back to them.

https://github.com/venkytv/botmand/assets/718613/5f59dcd6-7bd4-406d-b481-029fc1e3463b

### Installation

```
make install
```

This copies the bot's config as well as the bot script to the
default config directory: `~/botmand-engines`.

If `botmand` is already running, make it reload its engines:

```
pkill -HUP botmand
```

### Usage

Once installed, the bot is triggered by the a directed message saying "hello"
in any channel it is present in, at which point it starts a thread where it 
repeats anything anybody says there. Once this gets irritating enough, type
"bye" to terminate the bot.
