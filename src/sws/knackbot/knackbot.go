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
var ircmuc_jid = flag.String("ircmuc", "", "irc muc jid (required)")
var ircmuc_password = flag.String("ircmucpw", "", "irc muc password")

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
    if *muc_jid == "" || *ircmuc_jid == "" || *nick == "" || *nick == "knackbot" {
        flag.Usage()
    }

    if !*notls {
		xmpp.DefaultConfig = tls.Config{
			ServerName:         serverName(*server),
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
                if bareJid(v.Remote) == *ircmuc_jid {
                    if nickName(v.Remote) == *nick {
                        fmt.Println("Was own nick, not forwarding...")
                    } else {
                        fmt.Println("Was irc muc, forwarding to ", *muc_jid)
						talk.SendPresence(xmpp.Presence{From: (*muc_jid) + "/knackbot", To: (*muc_jid) + "/" + nickName(v.Remote)})
                        talk.Send(xmpp.Chat{Remote: *muc_jid, Type: "groupchat", Text: nickName(v.Remote) + " said: " + v.Text})
						talk.SendPresence(xmpp.Presence{To: (*muc_jid) + "/knackbot", From: (*muc_jid) + "/" + nickName(v.Remote)})
                    }
                }
                if bareJid(v.Remote) == *muc_jid {
                    fmt.Println("Was muc from nick: ", nickName(v.Remote))
                    if nickName(v.Remote) == *nick {
                        fmt.Println("Forward message to irc muc...")
                        talk.Send(xmpp.Chat{Remote: *ircmuc_jid, Type: "groupchat", Text: v.Text})
                    }
                }
			case xmpp.Presence:
				fmt.Println("presence. from:", v.From, "show:", v.Show, "type:", v.Type, "status:", v.Status)
			}
		}
	}()

	if *muc_password != "" {
		talk.JoinProtectedMUC(*muc_jid, "knackbot", *muc_password, xmpp.CharHistory, 0, nil)
	} else {
		talk.JoinMUC(*muc_jid, "knackbot", xmpp.CharHistory, 0, nil)
	}
	if *ircmuc_password != "" {
		talk.JoinProtectedMUC(*ircmuc_jid, *nick, *ircmuc_password, xmpp.CharHistory, 0, nil)
	} else {
		talk.JoinMUC(*ircmuc_jid, *nick, xmpp.CharHistory, 0, nil)
	}
    for {
		runtime.Gosched()
		time.Sleep(1 * time.Second)
	}
}