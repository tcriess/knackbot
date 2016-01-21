// Copyright 2013 Flo Lauber <dev@qatfy.at>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(flo):
//   - support password protected MUC rooms
//   - cleanup signatures of join/leave functions
package xmpp

import (
	"fmt"
	"time"
)

const (
	nsMUC     = "http://jabber.org/protocol/muc"
	nsMUCUser = "http://jabber.org/protocol/muc#user"
	NoHistory = 0
	CharHistory = 1
	StanzaHistory = 2
	SecondsHistory = 3
	SinceHistory = 4
)

func (c *Client) JoinMUCNoHistory(jid, nick string) {
	if nick == "" {
		nick = c.jid
	}
	fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n"+
		"<x xmlns='%s'>"+
		"<history maxchars='0'/></x>\n"+
		"</presence>",
		xmlEscape(jid), xmlEscape(nick), nsMUC)
}

// xep-0045 7.2
func (c *Client) JoinMUC(jid, nick string, history_type, history int, history_date *time.Time) {
	if nick == "" {
		nick = c.jid
	}
	switch history_type {
	case NoHistory:
		fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n" +
			"<x xmlns='%s' />\n" +
			"</presence>",
				xmlEscape(jid), xmlEscape(nick), nsMUC)
	case CharHistory:
		fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n" +
			"<x xmlns='%s'>\n" +
			"<history maxchars='%d'/></x>\n"+
			"</presence>",
				xmlEscape(jid), xmlEscape(nick), nsMUC, history)
	case StanzaHistory:
		fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n" +
			"<x xmlns='%s'>\n" +
			"<history maxstanzas='%d'/></x>\n"+
			"</presence>",
				xmlEscape(jid), xmlEscape(nick), nsMUC, history)
	case SecondsHistory:
		fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n" +
			"<x xmlns='%s'>\n" +
			"<history seconds='%d'/></x>\n"+
			"</presence>",
				xmlEscape(jid), xmlEscape(nick), nsMUC, history)
	case SinceHistory:
		if history_date != nil {
			fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n" +
				"<x xmlns='%s'>\n" +
				"<history since='%s'/></x>\n" +
				"</presence>",
					xmlEscape(jid), xmlEscape(nick), nsMUC, history_date.Format(time.RFC3339))
		}
	}
}

// xep-0045 7.2.6
func (c *Client) JoinProtectedMUC(jid, nick string, password string, history_type, history int, history_date *time.Time) {
	if nick == "" {
		nick = c.jid
	}
	switch history_type {
	case NoHistory:
		fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n" +
			"<x xmlns='%s'>\n" +
			"<password>%s</password>\n"+
			"</presence>",
				xmlEscape(jid), xmlEscape(nick), nsMUC, xmlEscape(password))
	case CharHistory:
		fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n" +
			"<x xmlns='%s'>\n" +
			"<password>%s</password>\n"+
			"<history maxchars='%d'/></x>\n"+
			"</presence>",
				xmlEscape(jid), xmlEscape(nick), nsMUC, xmlEscape(password), history)
	case StanzaHistory:
		fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n" +
			"<x xmlns='%s'>\n" +
			"<password>%s</password>\n"+
			"<history maxstanzas='%d'/></x>\n"+
			"</presence>",
				xmlEscape(jid), xmlEscape(nick), nsMUC, xmlEscape(password), history)
	case SecondsHistory:
		fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n" +
			"<x xmlns='%s'>\n" +
			"<password>%s</password>\n"+
			"<history seconds='%d'/></x>\n"+
			"</presence>",
				xmlEscape(jid), xmlEscape(nick), nsMUC, xmlEscape(password), history)
	case SinceHistory:
		if history_date != nil {
			fmt.Fprintf(c.conn, "<presence to='%s/%s'>\n" +
				"<x xmlns='%s'>\n" +
				"<password>%s</password>\n"+
				"<history since='%s'/></x>\n" +
				"</presence>",
					xmlEscape(jid), xmlEscape(nick), nsMUC, xmlEscape(password), history_date.Format(time.RFC3339))
		}
	}
}

// xep-0045 7.14
func (c *Client) LeaveMUC(jid string) {
	fmt.Fprintf(c.conn, "<presence from='%s' to='%s' type='unavailable' />",
		c.jid, xmlEscape(jid))
}
