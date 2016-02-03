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
	"time"
)

var server = flag.String("server", "", "server (<host> part of username)")
var username = flag.String("username", "", "username (<name>@<host>)")
var password = flag.String("password", "", "password")
var status = flag.String("status", "xa", "status")
var statusMessage = flag.String("status-msg", "I for one welcome our new codebot overlords.", "status message")
var notls = flag.Bool("notls", false, "No TLS")
var debug = flag.Bool("debug", false, "debug output")
var session = flag.Bool("session", false, "use server session")
var resource = flag.String("resource", "knackbot", "resource")
var muc_jid = flag.String("muc", "", "muc jid(s) (required), comma separated")
var muc_password = flag.String("mucpw", "", "muc password(s)")
var nick = flag.String("nick", "", "nick to use in irc. muc messages from this nick are forwarded to the irc muc, must be different from knackbot")

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
    if *muc_jid == "" || *nick == "" || *nick == "knackbot" {
        // || *ircmuc_jid == ""
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
                fmt.Println("Bare jid: " + bareJid(v.Remote))
                if bareJid(v.Remote) == *muc_jid {
                    fmt.Println("Was muc from nick: ", nickName(v.Remote))
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
    /*
	if *ircmuc_password != "" {
		talk.JoinProtectedMUC(*ircmuc_jid, *nick, *ircmuc_password, xmpp.CharHistory, 0, nil)
	} else {
		talk.JoinMUC(*ircmuc_jid, *nick, xmpp.CharHistory, 0, nil)
	}
	*/

    time.Sleep(2 * time.Second)

    talk.LeaveMUC(*muc_jid)

    time.Sleep(2 * time.Second)

}