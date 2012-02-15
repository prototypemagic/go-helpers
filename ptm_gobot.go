// Steve Phillips / elimisteve
// 2011.03.25

package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	BOT_NICK    = "ptm_gobot"
	IRC_SERVER  = "irc.freenode.net:6667"
	IRC_CHANNEL = "#ptmtest"
)

// Only one connection allowed (subject to change)
var conn net.Conn

func main() {
	var err os.Error

	conn, err = ircSetup()
	defer conn.Close()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		//
		// Main loop
		//
		read_buf := make([]byte, 512)
		for {
			length, err := conn.Read(read_buf)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			} else {
				msg := string(read_buf[:length])
				fmt.Printf("%v\n", msg)
				//
				// Respond to PING
				//
				if strings.HasPrefix(msg, "PING") {
					ircMsg("PONG " + msg)
				}
				//
				// Parse nick
				//
				if strings.Contains(msg, "PRIVMSG") {
					// Splits message in 2 parts.  Trailing [1:] ignores leading :
					nick := strings.SplitN(msg, "!", 2)[0][1:] // I love Go
					fmt.Printf("Sent by " + nick)
				}
				//
				// Parse msg
				//

				// parseQuit()
				// parseTime()
				// parseTime24()
				// parseProjectRead()
				// parseProjectEmpty()
				// parseProjectAdd()
				// parseGoogle()
				// parseG()
				// parseLucky()
				// parseWikilink()
				// parseMsg()
				// parsePrivMsg()
				// parseWiki()
				// parseHelp()
				// parseMembers()
				// parseStats()
				// parseUserlist()
				// parseIrcbot()
				// parseIrcbot()
				// parseEcho()
				// parseBadlist()
				// parseGoodlist()
				// parseProfanity()
				// parseShutup()
				// parseFuckyou()
				// parseDefine()
				// parseDefine()
				// parseUrbandef()
				// parseUrbandef()
				// parseWhoami()
				// parseMorse()
				// parseConvert()
			}
		}
	}
}

func ircSetup() (net.Conn, os.Error) {
	//IRC_CHANNELS := []string{"#ptmtest"}
	var err os.Error
	conn, err = net.Dial("tcp", IRC_SERVER)
	if err == nil {
		ircMsg("NICK " + BOT_NICK)
		ircMsg("USER " + strings.Repeat(BOT_NICK+" ", 4))
		ircMsg("JOIN " + IRC_CHANNEL)
		// for _, channel := range IRC_CHANNELS {
		//     ircMsg("JOIN " + channel)
		// }
	} else {
		fmt.Printf("Failed to connect.\n")
		os.Exit(1)
	}
	return conn, err
}

func ircMsg(str string) {
	conn.Write([]uint8(str + "\n"))
}
