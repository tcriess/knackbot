package main

import (
    //"bufio"
    "strings"
    "flag"
    "fmt"
    "os"
    "log"
    "crypto/tls"
	"github.com/mattn/go-xmpp"
	"runtime"
	"time"
	"github.com/thoj/go-ircevent"
)

var server = flag.String("server", "", "server")
var username = flag.String("username", "", "username")
var password = flag.String("password", "", "password")
var status = flag.String("status", "xa", "status")
var statusMessage = flag.String("status-msg", "I for one welcome our new codebot overlords.", "status message")
var notls = flag.Bool("notls", false, "No TLS")
var debug = flag.Bool("debug", false, "debug output")
var session = flag.Bool("session", false, "use server session")
var resource = flag.String("resource", "knackbot", "resource")
var muc_jid = flag.String("muc", "", "muc jid (required)")
var muc_password = flag.String("mucpw", "", "muc password")
var nick = flag.String("nick", "", "nick to use in irc muc (messages from this nick in the muc are forwarded to the irc muc), must be different from knackbot")
//var ircmuc_jid = flag.String("ircmuc", "", "irc muc jid (required)")
//var ircmuc_password = flag.String("ircmucpw", "", "irc muc password")
var ircurl = flag.String("ircurl", "", "irc host url (<host>:<port>)")
var ircchannel = flag.String("ircchannel", "", "irc channel to join")
var ircnick = flag.String("ircnick", "", "irc nick")

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

	if *username == "" || *password == "" {
		if *debug && *username == "" && *password == "" {
			fmt.Fprintf(os.Stderr, "no username or password were given; attempting ANONYMOUS auth\n")
		} else if *username != "" || *password != "" {
			flag.Usage()
		}
	}
    if *muc_jid == "" || *nick == "" || *nick == "knackbot" { // || *ircmuc_jid == ""
        flag.Usage()
    }


    if !*notls {
		xmpp.DefaultConfig = tls.Config{
			ServerName:         serverName(*server),
			InsecureSkipVerify: false,
		}
	}

	if(*ircurl == "" || *ircchannel == "" || *ircnick=="") {
		flag.Usage()
	}

    ircobj := irc.IRC(*ircnick, *ircnick) //Create new ircobj
    //Set options
    //ircobj.UseTLS = true //default is false
    //ircobj.TLSOptions //set ssl options
    //ircobj.Password = "[server password]"
    //Commands

/*
    ircobj.AddCallback("*", func(event *irc.Event) {
        //event.Message() contains the message
        //event.Nick Contains the sender
        //event.Arguments[0] Contains the channel
        fmt.Println("catchall: ", event.Code, ":", event.Source, ",", event.Nick, ",", event.Message())
    });
*/




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

    go func() {
		for {
			chat, err := talk.Recv()
			if err != nil {
				log.Fatal(err)
			}
			switch v := chat.(type) {
			case xmpp.Chat:
				fmt.Println(v.Remote, v.Text)
                fmt.Println("Bare jid: "+bareJid(v.Remote))
                /*
                if bareJid(v.Remote) == *ircmuc_jid {
                    if nickName(v.Remote) == *nick {
                        fmt.Println("Was own nick, not forwarding...")
                    } else if nickName(v.Remote) != "" {
                        fmt.Println("Was irc muc, forwarding to ", *muc_jid)
						talk.SendPresence(xmpp.Presence{From: (*muc_jid) + "/knackbot", To: (*muc_jid) + "/" + nickName(v.Remote)})
                        //talk.Send(xmpp.Chat{Remote: *muc_jid, Type: "groupchat", Text: nickName(v.Remote) + " said: " + v.Text})
                        talk.Send(xmpp.Chat{Remote: *muc_jid, Type: "groupchat", Text: v.Text})
						talk.SendPresence(xmpp.Presence{To: (*muc_jid) + "/knackbot", From: (*muc_jid) + "/" + nickName(v.Remote)})
                    }
                }
                */
                if bareJid(v.Remote) == *muc_jid {
                    fmt.Println("Was muc from nick: ", nickName(v.Remote))
                    if nickName(v.Remote) == *nick {
                        if strings.HasPrefix(v.Text, "knackbot: ") {
                            fmt.Println("Knackbot command:", v.Text)
                        } else {
                            //fmt.Println("Forward message to irc muc...")
                            //talk.Send(xmpp.Chat{Remote: *ircmuc_jid, Type: "groupchat", Text: v.Text})
                            fmt.Println("Forward message to irc...")
                            if ircobj.Connected() {
                                ircobj.Privmsg(*ircchannel, v.Text)
                                fmt.Println(*ircchannel, v.Text)
                            }
                        }
                    }
                }
			case xmpp.Presence:
				fmt.Println("presence. from:", v.From, "show:", v.Show, "type:", v.Type, "status:", v.Status)
			}
		}
	}()

    ircobj.AddCallback("001", func(event *irc.Event) {
        // now we can join.
        ircobj.Join(*ircchannel)
        //ircobj.SendRawf("NAMES %s", *ircchannel)
    })

    ircobj.AddCallback("353", func(event *irc.Event) {
        // RPL_NAMREPLY
        // NAMES response
        fmt.Println("353", event.Arguments)
        if len(event.Arguments) > 2 && event.Arguments[2] == *ircchannel {
            fmt.Println("List of nicks in channel:", event.Arguments[0], ", ", event.Code, ":", event.Source, ",", event.Nick, ",", event.Message())
            nicks := strings.Split(event.Message(), " ")
            fmt.Println(nicks)
            talk.Send(xmpp.Chat{Remote: *muc_jid, Type: "groupchat", Text: "Nicks in IRC: " + event.Message()})
        }
    });

    ircobj.AddCallback("332", func(event *irc.Event) {
        // RPL_TOPIC
        fmt.Println("332", event.Arguments)
        if len(event.Arguments) > 1 && event.Arguments[1] == *ircchannel {
            fmt.Println("Channel subject:", event.Code, ":", event.Source, ",", event.Nick, ",", event.Message())
        }
    })

    ircobj.AddCallback("PRIVMSG", func(event *irc.Event) {
        if event.Nick != *nick && event.Nick != "" {
            fmt.Println("Forward message to MUC")
            talk.SendPresence(xmpp.Presence{From: (*muc_jid) + "/knackbot", To: (*muc_jid) + "/" + event.Nick})
            //talk.Send(xmpp.Chat{Remote: *muc_jid, Type: "groupchat", Text: nickName(v.Remote) + " said: " + v.Text})
            talk.Send(xmpp.Chat{Remote: *muc_jid, Type: "groupchat", Text: event.Message()})
            talk.SendPresence(xmpp.Presence{To: (*muc_jid) + "/knackbot", From: (*muc_jid) + "/" + event.Nick})
        } else {
            fmt.Println("Was my nick - not forwarding")
        }
    })

    ircobj.Connect(*ircurl) //Connect to server

    defer ircobj.Disconnect()

    go ircobj.Loop()

	if *muc_password != "" {
		talk.JoinProtectedMUC(*muc_jid, "knackbot", *muc_password, xmpp.CharHistory, 0, nil)
	} else {
		talk.JoinMUC(*muc_jid, "knackbot", xmpp.CharHistory, 0, nil)
	}
    /*
	if *ircmuc_password != "" {
		talk.JoinProtectedMUC(*ircmuc_jid, *nick, *ircmuc_password, xmpp.CharHistory, 0, nil)
	} else {
		talk.JoinMUC(*ircmuc_jid, *nick, xmpp.CharHistory, 0, nil)
	}
	*/

    go func() {
        t := time.NewTicker(5 * time.Minute)
        for {
            ircobj.SendRawf("NAMES %s", *ircchannel)
            <-t.C
        }
    }()

    for {
		runtime.Gosched()
		time.Sleep(1 * time.Second)
	}
}