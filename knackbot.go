package main

import (
    "strings"
    "flag"
    "fmt"
    "os"
    "log"
    "crypto/tls"
    "runtime"
    "time"
    "github.com/thoj/go-ircevent"
    "github.com/tcriess/go-xmpp"
)

var maxProcs = flag.Int("procs", 1, "set GOMAXPROCS")
var server = flag.String("server", "", "server (<host> part of username)")
var username = flag.String("username", "", "username (<name>@<host>)")
var password = flag.String("password", "", "password")
var status = flag.String("status", "xa", "status")
var statusMessage = flag.String("status-msg", "I for one welcome our new codebot overlords.", "status message")
var notls = flag.Bool("notls", false, "No TLS")
var ircnotls = flag.Bool("ircnotls", false, "No TLS for irc")
var debug = flag.Bool("debug", false, "debug output")
var session = flag.Bool("session", false, "use server session")
var resource = flag.String("resource", "knackbot", "resource")
var muc_jid = flag.String("muc", "", "muc jid(s), comma-separated (required)")
var muc_password = flag.String("mucpw", "", "muc password(s)")
var nick = flag.String("nick", "", "nick to use in irc (messages from this nick in the muc are forwarded to irc), must be different from knackbot")
var ircurl = flag.String("ircurl", "", "irc host url (<host>:<port>)")
var ircchannel = flag.String("ircchannel", "", "irc channel(s) to join")
var ircserver = flag.String("ircserver", "", "irc server name (<host> part of ircurl)")

func serverName(host string) string {
    return strings.Split(host, ":")[0]
}

func nickName(jid string) string {
    var parts = strings.Split(jid, "/")
    if len(parts) > 1 {
        return parts[1]
    } else {
        return ""
    }
}

func bareJid(jid string) string {
    return strings.Split(jid, "/")[0]
}

func main() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "usage: knackbot [options]\n")
        flag.PrintDefaults()
        os.Exit(2)
    }
    flag.Parse()

    runtime.GOMAXPROCS(*maxProcs)

    if *username == "" || *password == "" {
        if *debug && *username == "" && *password == "" {
            fmt.Fprintf(os.Stderr, "no username or password were given; attempting ANONYMOUS auth\n")
        } else if *username != "" || *password != "" {
            flag.Usage()
        }
    }
    if *muc_jid == "" || *nick == "" || *nick == "knackbot" {
        flag.Usage()
    }

    if !*notls {
        xmpp.DefaultConfig = tls.Config{
            ServerName:         serverName(*server),
            InsecureSkipVerify: false,
        }
    }

    if (*ircurl == "" || *ircchannel == "") {
        flag.Usage()
    }

    ircchannels := strings.Split(*ircchannel, ",")
    muc_jids := strings.Split(*muc_jid, ",")
    muc_passwords := strings.Split(*muc_password, ",")

    if len(ircchannels) != len(muc_jids) || len(muc_jids) < len(muc_passwords) {
        flag.Usage()
    }

    if len(muc_passwords) < len(muc_jids) {
        for i := len(muc_passwords); i < len(muc_jids); i++ {
            muc_passwords = append(muc_passwords, "")
        }
    }
    muc_jids_channels := make(map[string]string)
    channels_muc_jids := make(map[string]string)
    muc_jids_pws := make(map[string]string)
    for i, mj := range muc_jids {
        muc_jids_pws[mj] = muc_passwords[i]
        muc_jids_channels[mj] = ircchannels[i]
    }
    for i, c := range ircchannels {
        channels_muc_jids[c] = muc_jids[i]
    }

    ircobj := irc.IRC(*nick, *nick) //Create new ircobj
    if !*ircnotls {
        if *ircserver == "" {
            flag.Usage()
        }
        ircobj.UseTLS = true //default is false
        ircobj.TLSConfig = &tls.Config{//set ssl options
            ServerName:         serverName(*ircserver),
            InsecureSkipVerify: false,
        }
    }

    var talk *xmpp.Client
    var err error
    options := xmpp.Options{Host: *server,
        User:           *username,
        Password:       *password,
        NoTLS:          *notls,
        Debug:          *debug,
        Session:        *session,
        Status:         *status,
        StatusMessage:  *statusMessage,
        Resource:       *resource,
    }

    talk, err = options.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    defer talk.Close()

    MUC2Channel := func(jid string) string {
        if channel, ok := muc_jids_channels[jid]; ok {
            return channel
        }
        return ""
    }

    Channel2MUC := func(channel string) string {
        if muc_jid, ok := channels_muc_jids[channel]; ok {
            return muc_jid
        }
        return ""
    }

    go func() {
        for {
            chat, err := talk.Recv()
            if err != nil {
                log.Fatal(err)
            }
            switch v := chat.(type) {
            case xmpp.Chat:
                go func(rem, text string) {
                    log.Println(rem, text)
                    log.Println("Bare jid: " + bareJid(rem))
                    barejid := bareJid(rem)
                    if channel := MUC2Channel(barejid); channel != "" {
                        log.Println("Was muc from nick: ", nickName(rem))
                        if nickName(rem) == *nick {
                            if strings.HasPrefix(strings.ToLower(text), "kb ") {
                                log.Println("Knackbot command:", text)
                                parts := strings.SplitN(strings.ToLower(text), " ", 2)
                                log.Println(parts)
                                if len(parts) > 1 {
                                    switch parts[1] {
                                    case "names":
                                        ircobj.SendRawf("NAMES %s", channel)
                                    case "topic":
                                        ircobj.SendRawf("TOPIC %s", channel)
                                    case "status":
                                        ircconnected := ircobj.Connected()
                                        if ircconnected {
                                            talk.Send(xmpp.Chat{Remote: barejid, Type: "groupchat", Text: "IRC connected"})
                                        } else {
                                            talk.Send(xmpp.Chat{Remote: barejid, Type: "groupchat", Text: "IRC not connected"})
                                        }
                                    default:
                                        log.Println("Unknown command", parts[1])
                                    }
                                }
                            } else {
                                log.Println("Forward message to irc...")
                                if ircobj.Connected() {
                                    ircobj.Privmsg(channel, text)
                                    log.Println(channel, text)
                                } else {
                                    log.Println("Could not foward to channel", channel, "- not connected")
                                    _, err := talk.Send(xmpp.Chat{Remote: barejid, Type: "groupchat", Text: "Could not foward to channel" + channel + " - not connected"})
                                    if err != nil {
                                        log.Fatal(err)
                                    }
                                }
                            }
                        }
                    }
                }(v.Remote, v.Text)
            case xmpp.Presence:
                log.Println("presence. from:", v.From, "show:", v.Show, "type:", v.Type, "status:", v.Status)
            }
        }
    }()

    ircobj.AddCallback("001", func(event *irc.Event) {
        go func() {
            // now we can join.
            for _, channel := range ircchannels {
                ircobj.Join(channel)
            }
        }()
    })

    ircobj.AddCallback("353", func(event *irc.Event) {
        // RPL_NAMREPLY
        // NAMES response
        go func(arguments []string, source, code, nick, message string) {
            log.Println("353", arguments)
            if len(arguments) > 2 {
                if jid := Channel2MUC(arguments[2]); jid != "" {
                    log.Println("List of nicks in channel:", arguments[2], arguments[0], ", ", code, ":", source, ",", nick, ",", message)
                    nicks := strings.Split(message, " ")
                    log.Println(nicks)
                    _, err := talk.Send(xmpp.Chat{Remote: jid, Type: "groupchat", Text: "Nicks in IRC channel " + arguments[2] + ": " + message})
                    if err != nil {
                        log.Fatal(err)
                    }
                }
            }
        }(event.Arguments, event.Source, event.Code, event.Nick, event.Message())
    });

    ircobj.AddCallback("332", func(event *irc.Event) {
        // RPL_TOPIC
        go func(arguments []string, source, code, nick, message string) {
            log.Println("332", arguments)
            if len(arguments) > 1 {
                if jid := Channel2MUC(arguments[1]); jid != "" {
                    log.Println("Channel subject:", code, ":", source, ",", nick, ",", message)
                    _, err := talk.SendTopic(xmpp.Chat{Remote: jid, Type: "groupchat", Text: message})
                    if err != nil {
                        log.Fatal(err)
                    }
                }
            }
        }(event.Arguments, event.Source, event.Code, event.Nick, event.Message())
    })

    ircobj.AddCallback("PRIVMSG", func(event *irc.Event) {
        go func(arguments []string, source, code, enick, message string) {
            log.Println("Got privmsg event:", "nick:", enick, "source:", source)
            if enick != *nick && enick != "" && len(arguments) > 0 {
                if jid := Channel2MUC(arguments[0]); jid != "" {
                    log.Println("Forward message to MUC")
                    _, err := talk.SendPresence(xmpp.Presence{From: jid + "/knackbot", To: jid + "/" + enick})
                    if err != nil {
                        log.Fatal(err)
                    }
                    _, err = talk.Send(xmpp.Chat{Remote: jid, Type: "groupchat", Text: message})
                    if err != nil {
                        log.Fatal(err)
                    }
                    _, err = talk.SendPresence(xmpp.Presence{To: jid + "/knackbot", From: jid + "/" + enick})
                    if err != nil {
                        log.Fatal(err)
                    }
                }
            } else {
                log.Println("Was my nick - not forwarding")
            }
        }(event.Arguments, event.Source, event.Code, event.Nick, event.Message())
    })

    ircobj.Connect(*ircurl) //Connect to server

    defer ircobj.Disconnect()

    go ircobj.Loop()

    for jid, pw := range muc_jids_pws {
        if pw != "" {
            _, err := talk.JoinProtectedMUC(jid, "knackbot", pw, xmpp.CharHistory, 0, nil)
            if err != nil {
                log.Fatal(err)
            }
        } else {
            _, err := talk.JoinMUC(jid, "knackbot", xmpp.CharHistory, 0, nil)
            if err != nil {
                log.Fatal(err)
            }
        }
    }

    // get current nicks and topic once per hour
    go func() {
        t := time.NewTicker(60 * time.Minute)
        for {
            <-t.C
            for _, channel := range ircchannels {
                ircobj.SendRawf("NAMES %s", channel)
                ircobj.SendRawf("TOPIC %s", channel)
            }
        }
    }()

    for {
        runtime.Gosched()
        time.Sleep(1 * time.Second)
    }
}