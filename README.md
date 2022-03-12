# wvlist
WVList is a Werke Verzeichnis database available over HTTP(s) with community submissions. It is being hosted by the creator at [wvlist.net](https://wvlist.net) [(onion service)](http://xtjuwqe4hlqojbknh2kwlm5nh4usj2yh5ceuh2x4gy27wirurxh3qoid.onion/), and can be hosted by anyone else as well. This project is intended to be a reference tool for the catalogue complete works list of composers and the numbering list used to identify each composition of a composer, even when ambiguous titles are used. Well-known examples of WV catalogues are [Bach Werke Verzeichnis](https://www.bach-cantatas.com/Bach-Werke-Verzeichnis.pdf), [K numbers used to classify Mozart compositions](https://www.mozartproject.org/what-are-mozart-k-numbers/), and using Opus numbers to classify published compositions.

# Dependencies
In addition to the modules imported by the Go compiler, this project requires the Lilypond music engraving software version >2.20.0. This can be installed in many ways. Instructions for how to install Lilypond can be found [here](https://lilypond.org/download.html).

# Installation
- Install Lilypond in your preferred way. Refer to [Lilypond installation instructions for compiling from source or installing in another way](https://lilypond.org/download.html).
- Clone the repository into your VPS, and change directory into the home directory of this project.

`git clone https://github.com/FiskFan1999/wvlist.git && cd wvlist`
- Highly recommended: checkout to the latest release instead of unstable HEAD
- Copy the default configuration file.

`cp config.default.json config.json`
- Change the configuration settings as required. (Refer to `config.txt` to describe each )
- Compile the code via `make`
- Run wvlist via `./wvlist run` Refer to instructions about flags (coming soon).

# Chat
Please feel free to join the conversation on IRC.

Main network:
- [`irc.ergo.chat:6697 (TLS) #wvlist`](https://ergo.chat/kiwi/#wvlist) (Click for browser client)

Alternate networks:
- `irc.oftc.net #wvlist`
- `irc.libera.chat #wvlist-dev`
- `irc.williamrehwinkel.net #wvlist`

Available on Matrix: [`#wvlist-dev:libera.chat`](https://matrix.to/#/#wvlist-dev:libera.chat)
