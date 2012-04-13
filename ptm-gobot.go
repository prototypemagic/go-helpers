// Steve Phillips / elimisteve
// 2011.03.25

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	IRC_SERVER      = "irc.freenode.net:6667"
	BOT_NICK        = "ptm_gobot"
	IRC_CHANNEL     = "#prototypemagic"
	PREFACE         = "PRIVMSG " + IRC_CHANNEL + " :"

	REPO_BASE_PATH  = "/home/steve/django_projects/"
	// REPO_BASE_PATH  = "/home/ubuntu/django_projects/"
	REPO_INDEX_FILE = ".index"
	GIT_PORT        = "6666"
	WEBHOOK_PORT    = "7777"
)

type GitCommit struct {
	Author string
	Email string
	Repo string
	Message string
}

func checkError(where string, err error) {
	if err != nil {
		log.Fatalf(where + ": " + err.Error())
	}
}

// Only one connection allowed (subject to change)
var conn net.Conn

// Anything passed to this channel is echoed into IRC_CHANNEL
var irc = make(chan string)

func main() {
	// If the constants are valid, this program cannot crash. Period.
	defer func() {
		if err := recover(); err != nil {
			msg := fmt.Sprintf("Recovered from nasty error in main: %v\n", err)
			// ircMsg(msg)
			fmt.Print(msg)
		}
	}()

	// Connect to IRC
	conn = ircSetup()
	defer conn.Close()

	// Listen for repo names on port GIT_PORT, then echo info
	// from latest commit into IRC_CHANNEL. Currently triggered by
	// post-receive git hooks.
	go gitListener()

	// Listen for (WebHook-powered) JSON POSTs from GitHub to port
	// WEBHOOK_PORT
	go webhookListener()

	// Anything passed to the `irc` channel (get it?) is echoed into
	// IRC_CHANNEL
	go func() {
		for { ircMsg(<-irc) }
	}()

	//
	// Main loop
	//
	read_buf := make([]byte, 512)
	for {
		// n bytes read
		n, err := conn.Read(read_buf)
		checkError("conn.Read", err)
		data := string(read_buf[:n-2])  // Ignore trailing \r\n
		fmt.Printf("%v\n", data)
		//
		// Respond to PING
		//
		if strings.HasPrefix(data, "PING") {
			rawIrcMsg("PONG " + data)
		}
		//
		// Parse nick, msg
		//

		// Avoids ~global var risk by resetting these to "" each loop
		var msg, nick string = "", ""

		if strings.Contains(data, "PRIVMSG") {
			// structure of `data` == :nick!host PRIVMSG #channel :msg

			// nick == everything after first char, before first !
			nick = strings.SplitN(data[1:], "!", 2)[0]
			fmt.Printf("Nick: '%v'\n", nick)

			// msg == everything after second :
			msg = strings.SplitN(data, ":", 3)[2]
			fmt.Printf("Message: '%v'\n", msg)
		}
		//
		// ADD YOUR CODE (or function calls) HERE
		//
	}
}

func ircSetup() net.Conn {
	var err error
	// Avoid the temptation... `conn, err := ...` silently shadows the
	// global `conn` variable!
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

func repoToNonMergeCommit(repoPath, repoName string) GitCommit {
	// defer func(where string) {
	// 	if err := recover(); err != nil {
	// 		msg := fmt.Sprintf("Recovered from error in %v: %v",
	// 			where, err)
	// 		ircMsg(msg)
	// 		log.Print(msg)
	// 	}
	// }("repoToNonMergeCommit")
	defer func() {
		if err := recover(); err != nil {
			msg := fmt.Sprintf("Recovered from error in repoToNonMergeCommit: %v",
				err)
			ircMsg(msg)
			log.Print(msg)
		}
	}()

	GIT_COMMAND := "git log -1"
	GIT_COMMAND_IF_MERGE := "git log -2"

	output := gitCommandToOutput(repoPath, GIT_COMMAND)
	if output == "" {
		return GitCommit{}
	}
	fmt.Printf("original output: \n%s\n", output)
	if strings.Contains(output, "\nMerge:") {
		commitStr := gitCommandToOutput(repoPath, GIT_COMMAND_IF_MERGE)
		regexStr := `commit [0-9a-f]{40}`
		getCommitLine := regexp.MustCompile(regexStr)
		commitStrs := getCommitLine.FindAllString(commitStr, -1)

		// Assumes at least one commit exists in output
		splitOnMe := commitStrs[len(commitStrs)-1]
		commits := strings.Split(commitStr, splitOnMe)
		
		output = commitStrs[0] + commits[len(commits)-1]
	}
	fmt.Printf("final output: \n%s\n", output)
	lines := strings.SplitN(output, "\n", 4)
	// commitLine := lines[0]
	authorLine := lines[1]
	// dateLine := lines[2]
	// TODO: Assumes entire commit message is on on one line
	commitMsg := strings.Replace(lines[3], "\n", "", -1)[4:]

	tokens := strings.Split(authorLine[8:], " ")
	authorEmail := tokens[len(tokens)-1]
	authorNames := tokens[:len(tokens)-1]
	// authorFirst := authorNames[0]
	author := strings.Join(authorNames, " ")
	commit := GitCommit {
	Author: author,
	Email: authorEmail,
	Repo: repoName,
	Message: commitMsg,
	}
	return commit
}


func gitCommandToOutput(repoPath, command string) string {
    args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = repoPath  // Where cmd is run from
	output, err := cmd.Output()
	if err != nil {
		log.Printf("repo not found at '%v'", repoPath)
		// Search ${repoName}/bare then ${repoName}_site for desired repo

		// Try bare/ if neither has been tried
		// (Should work on server)
		if !strings.HasSuffix(repoPath, "/bare") &&
			!strings.HasSuffix(repoPath, "_site") {
			return gitCommandToOutput(repoPath + "/bare", command)
		}
		// Try _site/ after trying bare/
		// (Usually necessary locally)
		if !strings.HasSuffix(repoPath, "_site") {
			// Ignore last len("/bare") chars
			repoPath = repoPath[:len(repoPath)-len("/bare")]
			return gitCommandToOutput(repoPath + "_site", command)
		}

		// Now both /bare and _site have been tried. Giving up.

		if repos := listRepos(); repos != "" {
			ircMsg("Repo not found. Options (probably): " + repos)
		}
		return ""
	}
	return string(output)
}


func gitRepoNameToPath(repoName string) string {
	// Alternatives to this I can think of: change global
	// REPO_BASE_PATH from const into var (global variables == bad!)
	repoBase := REPO_BASE_PATH[:]

	// If repoName begins with /, user is giving absolute path to repo
	if strings.HasPrefix(repoName, "/") {
		repoBase = ""
	}
    return repoBase + repoName
}


func gitRepoDataParser(repoName string) GitCommit {
	defer func() {
		if err := recover(); err != nil {
			msg := fmt.Sprintf("Recovered from error in gitRepoDataParser: %v",
				err)
			ircMsg(msg)
			log.Print(msg)
		}
	}()
	repoPath := gitRepoNameToPath(repoName)
	commit := repoToNonMergeCommit(repoPath, repoName)
	return commit
}

// gitListener listens on localhost:GIT_PORT for git repo names
// coming from git post-receive hooks, then
func gitListener() {
	defer func() {
		if err := recover(); err != nil {
			msg := fmt.Sprintf("Recovered from error in gitListener: %v",
				err)
			ircMsg(msg)
		}
	}()

	// Create TCP connection
	addr := "127.0.0.1:" + GIT_PORT
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

		gitCommit := gitRepoDataParser(repoName)
		empty := GitCommit{}
		if gitCommit != empty {
			msg := fmt.Sprintf(`%v pushed to %v: "%v"`,
				gitCommit.Author, repoName, gitCommit.Message)
			ircMsg(msg)
		}
		conn.Close()
	}
	return
}

func listRepos() string {
	result, err := ioutil.ReadFile(REPO_BASE_PATH + REPO_INDEX_FILE)
	if err != nil {
		fmt.Printf("%v\n", "No index file found")

		// List directory contents instead

		// TODO: Look for subdirectories of REPO_BASE_PATH containing
		// .git/ -- wait for it -- subdirectories

		// Only include directory names
		// TODO: Assumes sub-directories do not include spaces
		// cmd := exec.Command("ls -l | grep ^d | awk '{print $9}'")

		cmd := exec.Command("ls")
		cmd.Dir = REPO_BASE_PATH
		// If no error, `result` used after this block
		result, err = cmd.Output()
		if err != nil {
			fmt.Printf("%v\n", err)
			msg := "Can't run 'ls'?! Somebody screwed up REPO_BASE_PATH..."
			fmt.Printf("%v\n", msg)
			ircMsg(msg)
			return ""
		}
	}
	// `result` came from .index file or REPO_BASE_PATH dir listing
	repoStr := string(result)

	// No valid repo supplied, no .index file, and REPO_BASE_PATH
	// doesn't exist(?). TODO: double-check when it's not 6:30am after
	// you've stayed up all night...
	if repoStr == "" {
		return ""
	}

	repoNames := strings.Split(repoStr, "\n")

	// Remove last repo in repoNames if it's empty (comes from
	// trailing newline in REPO_INDEX_FILE)
	if repoNames[len(repoNames)-1] == "" {
		repoNames = repoNames[:len(repoNames)-1]
	}
	return strings.Join(repoNames, ", ")
}

//
// Accept GitHub Webhook data (qua JSON file) on port 7777
//

// TODO: Define JSON "template" as struct to capture its structure?

func webhookHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		info := fmt.Sprintf("req:\n\n%+#v\n", req)
		fmt.Printf("%v\n", info)
		// w.Write([]byte(info))
		return
	}
	body := make([]byte, req.ContentLength)
	_, err := req.Body.Read(body)
	defer req.Body.Close()
	if err != nil {
		fmt.Printf("Error in webhookHandler: %v\n", err)
		return
	}
	// fmt.Printf("Everything:\n%+#v\n\n", req)

	// fmt.Printf("Entire body:\n%s\n", body)

	decoded, err := url.Parse(fmt.Sprintf("%s", body))
	if err != nil {
		fmt.Printf("Error parsing body: %v\n", err)
	}
	decodedBody := decoded.Path

	// FIXME: Get req.FormValue("payload") or similar to work and
	// strip out the following bullshit...

	get_pusher := regexp.MustCompile(`"pusher":{"name":"(.*)","email`)
	str := get_pusher.FindStringSubmatch(decodedBody)[1]
	decodedURL, _ := url.Parse(str)
	quote := strings.Index(decodedURL.Path, `"`)
	author := decodedURL.Path[:quote]
	
	get_repo_name := regexp.MustCompile(`"repository":{"name":"(.*)","size`)
	str = get_repo_name.FindStringSubmatch(decodedBody)[1]
	decodedURL, _ = url.Parse(str)
	repo := decodedURL.Path

	irc <- fmt.Sprintf("%v just pushed to %v on GitHub!", author, repo)

	// See http://blog.golang.org/2011/01/json-and-go.html

	return

	// w.Write(append( []byte(fmt.Sprintf("%+#v\n\nBODY:\n%s\n\nREQUESTURI:\n%+#v",
	// 	req, body, req.RequestURI)) ))

	// // Echo body right back
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(body)
	// w.Header().Set("Content-Type", "text/plain")
	// w.Write([]byte(""))
}

// webhookListener listens on WEBHOOK_PORT (default: 7777) for JSON
// HTTP POSTs, parses the relevant data, then sends it over a channel
// to a function waiting for strings to echo into IRC_CHANNEL
func webhookListener() {
	http.HandleFunc("/webhook", webhookHandler)
	// err := http.ListenAndServeTLS(":"+WEBHOOK_PORT, "cert.pem", "key.pem", nil)
	err := http.ListenAndServe(":"+WEBHOOK_PORT, nil)
	if err != nil {
		fmt.Printf("Error in webhookListener ListenAndServe: %v\n", err)
		time.Sleep(2e9)
	}
}
