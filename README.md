IRC-to-MUC bridging bot
=======================

Installation
------------

Set your GOPATH environment variable to your favourite location. Then fetch the source:

    go get github.com/tcriess/knackbot
    
To install into $GOPATH/bin, do

    cd $GOPATH/src/github.com/tcriess/knackbot
    go install
  
Usage
-----

To use the bot, create a dedicated Jabber account. Make sure the bot has access to a Jabber MUC room.

    $GOPATH/bin/knackbot -server <jabber-server> -username <jabber-username> -password <jabber-password> -muc <jaber-muc-jids> -nick <nickname> -ircurl <IRC-server:port> -ircchannel <IRC-channels> -ircserver <IRC-server>

Then bot connects to the IRC server, using the provided nick, joins the provided channel(s) and it joins the provided Jabber MUC(s), using the nick "knackbot".
Whenever a message in one of the IRC channel is received, the message is forwarded to the corresponding Jabber MUC (the bot's Jabber nick is temporarily changed to match the source of the message).
Whenever a message in one of the MUCs is received *and the sender's nick is the one the bot uses in IRC*, the message is forwarded to the corresponsing IRC channel, with the exception of bot commands. These commands are "kb names", "kb topic" and "kb status". "kb names" lists the nicknames in the IRC channel, "kb topic" returns the channel topic and "kb status" indicates if the bot is connected to the IRC server.

To use the bot for multiple IRC channels (and the corresponding number of MUCs), use a comma-separated list in the "-muc" and "-ircchannel" arguments.