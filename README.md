# ProtoType Magic's Go helper functions

## Errors

1. Disconnection

    ERROR :Closing Link: ec2-23-21-53-88.compute-1.amazonaws.com (Ping timeout: 265 seconds)
    2012/04/30 18:08:27 conn.Read: EOF
    exit status 1


2. False Type Assertion

Looks like JSON templates aren't just convenient, but can aid in correctness:

    http: panic serving 50.57.128.197:45137: interface conversion: interface is nil, not map[string]interface {}
    /usr/local/go/src/pkg/net/http/server.go:576 (0x46deed)
    /tmp/bindist046461602/go/src/pkg/runtime/proc.c:1443 (0x41389b)
    /tmp/bindist046461602/go/src/pkg/runtime/iface.c:297 (0x40ae62)
    /tmp/bindist046461602/go/src/pkg/runtime/iface.c:285 (0x40adef)
    /home/ubuntu/github_repos/go-helpers/ptm-gobot.go:572 (0x40415f)
    webhookDataToGitCommit: head_commit := v.(map[string]interface{})
    /home/ubuntu/github_repos/go-helpers/ptm-gobot.go:520 (0x40395a)
    webhookHandler: commit := webhookDataToGitCommit(data)
    /usr/local/go/src/pkg/net/http/server.go:690 (0x461f52)
    /usr/local/go/src/pkg/net/http/server.go:924 (0x462dd4)
    /usr/local/go/src/pkg/net/http/server.go:656 (0x461d65)
    /tmp/bindist046461602/go/src/pkg/runtime/proc.c:271 (0x4119a1)


## ptm-gobot.go TODO

* Listen on GTalk for instructions

* Listen via the command line for instructions
