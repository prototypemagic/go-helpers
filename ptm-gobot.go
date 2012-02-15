// Steve Phillips / elimisteve
// 2011.03.25

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

const (
	IRC_SERVER      = "irc.freenode.net:6667"
	BOT_NICK        = "ptm-gobot"
	IRC_CHANNEL     = "#prototypemagic"
	PREFACE         = "PRIVMSG " + IRC_CHANNEL + " :"

	// REPO_BASE_PATH  = "/home/steve/django_projects/"
	REPO_BASE_PATH  = "/home/ubuntu/django_projects/"
	REPO_INDEX_FILE = ".index"
	GIT_LISTEN_PORT = "6666"
)

func checkError(where string, err error) {
	if err != nil {
		log.Fatalf(where + ": " + err.Error())
	}
}

// Only one connection allowed (subject to change)
var conn net.Conn

func main() {
	conn = ircSetup()
	defer conn.Close()

	// Listen for repo names on port GIT_LISTEN_PORT, then echo info
	// from latest commit into IRC_CHANNEL. Currently triggered by
	// post-receive git hooks.
	go gitListener()

	//
	// Main loop
	//
	read_buf := make([]byte, 512)
	for {
		length, err := conn.Read(read_buf)
		checkError("conn.Read", err)
		msg := string(read_buf[:length])
		fmt.Printf("%v\n", msg)
		//
		// Respond to PING
		//
		if strings.HasPrefix(msg, "PING") {
			rawIrcMsg("PONG " + msg)
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

func ircSetup() net.Conn {
	//IRC_CHANNELS := []string{"#prototypemagic"}
	var err error
	conn, err = net.Dial("tcp", IRC_SERVER)
	checkError("net.Dial", err)

	rawIrcMsg("NICK " + BOT_NICK)
	rawIrcMsg("USER " + strings.Repeat(BOT_NICK+" ", 4))
	rawIrcMsg("JOIN " + IRC_CHANNEL)
	return conn
}

// rawIrcMsg takes a string and writes it to the global TCP connection
// to the IRC server _verbatim_
func rawIrcMsg(str string) {
	conn.Write([]uint8(str + "\n"))
}

// ircMsg is a helper function that wraps rawIrcMsg, prefacing each
// message with PREFACE (usually `PRIVMSG $IRC_CHANNEL `)
func ircMsg(msg string) {
	rawIrcMsg(PREFACE + msg)
}

func gitRepoDataParser(repoName string) map[string]string {
	defer func() {
		if err := recover(); err != nil {
			msg := fmt.Sprintf("Recovered from error in gitRepoDataParser: %v",
				err)
			ircMsg(msg)
			log.Print(msg)
		}
	}()

	const (
		// REPO_BASE_PATH = "/home/ubuntu/django_projects/"
		GIT_COMMAND = "git log -1"
	)
	repoBase := string(REPO_BASE_PATH)

	// If given repo name begins with /, treat as user giving absolute
	// path
	if strings.HasPrefix(repoName, "/") {
		repoBase = ""
	}
	repoPath := repoBase + repoName
	// fmt.Printf("repoPath == %v\n", repoPath)

	// Create *Command object
	args := strings.Split(GIT_COMMAND, " ")
	// fmt.Printf("args == %v\n", args)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = repoPath  // Where cmd is run from
	output, err := cmd.Output()
	if err != nil {
		// ircMsg(fmt.Sprintf("Error from cmd.Output() in gitRepoDataParser: %v",
		// 	err))
		ircMsg("Invalid repo name. Options: " + listRepos())
		return nil
	}

	// `output` now contains GIT_COMMAND output

	lines := strings.SplitN(string(output), "\n", 4)
	// commitLine := lines[0]
	authorLine := lines[1]
	// dateLine := lines[2]
	commitMsg := strings.Replace(lines[3], "\n", "", -1)[4:]

	tokens := strings.Split(authorLine[8:], " ")
	authorEmail := tokens[len(tokens)-1]
	authorNames := tokens[:len(tokens)-1]
	// authorFirst := authorNames[0]
	author := strings.Join(authorNames, " ")

	repoInfo := map[string]string{}
	repoInfo["author"] = author
	repoInfo["email"] = authorEmail
	repoInfo["repo"] = repoName
	repoInfo["message"] = commitMsg
	return repoInfo
}

// gitListener listens on localhost:GIT_LISTEN_PORT for git repo names
// coming from git post-receive hooks, then
func gitListener() {
	defer func() {
		if err := recover(); err != nil {
			msg := fmt.Sprintf("Recovered from error in gitListener: %v",
				err)
			ircMsg(msg)
		}
	}()

	addr := "127.0.0.1:" + GIT_LISTEN_PORT
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	checkError("net.ResolveTCPAddr", err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError("net.ListenTCP", err)

	for {
		var buf [512]byte

		conn, err := listener.Accept()
		if err != nil {
			msg := fmt.Sprintf("Error in gitListener: %v\n", err)
			log.Printf(msg)
			ircMsg(msg)
			time.Sleep(2e9)
			continue
		}
		// Read n bytes
		n, err := conn.Read(buf[:])

		repoName := string(buf[:n])
		log.Printf("Repo name received: '%v'", repoName)
		// repoName = ""

		repoInfo := gitRepoDataParser(repoName)
		if repoInfo != nil {
			msg := fmt.Sprintf(`%v pushed to %v: "%v"`,
				repoInfo["author"], repoName, repoInfo["message"])
			ircMsg(msg)
		}
		conn.Close()
	}
	return
}

func listRepos() string {
	result, err := ioutil.ReadFile(REPO_BASE_PATH + REPO_INDEX_FILE)
	if err != nil {
		msg := fmt.Sprintf("No index file found")
		fmt.Printf("%v\n", msg)
		// ircMsg(msg)

		// List directory contents instead

		// TODO: Look for subdirectories of REPO_BASE_PATH containing
		// .git/, uh, subdirectories
		cmd := exec.Command("ls")
		cmd.Dir = REPO_BASE_PATH
		// If no error, `result` used after this block
		result, err = cmd.Output()
		if err != nil {
			msg = "Can't run 'ls'?! Somebody screwed up the REPO_BASE_PATH..."
			fmt.Printf("%v\n", msg)
			ircMsg(msg)
			return ""
		}
	}
	// `result` came from .index file or REPO_BASE_PATH dir listing
	repoNames := strings.Split(string(result), "\n")
	return strings.Join(repoNames, ", ")
}
