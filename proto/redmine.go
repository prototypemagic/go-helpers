// Steve Phillips / elimisteve
// 2011.05.27

package proto

import (
    "fmt"
    // "io/ioutil"
	"log"
    "net/http"
    "strings"
)

const (
	REDMINE_JSON_URL = "http://prototypemagic.com:3000/issues.json"
	BOT_REDMINE_KEY  = "ce734fe0d3fb6a47015d96940949d42d3ed0e4be"
)

// Maps stated assignee's name to assigned_to_id
var assigneeToID = map[string]string {
	"aj": "10",
	"carter": "13",
	"gobot": "20",
	"jay": "4",
	"jim": "17",
	"steve": "3",
}

// Maps ticketType to tracker_id
var trackerID = map[string]string {
	"bug": "1",
	"feature": "2",
}

func CreateTicket(project, ticketType, subject, description, assignee string) *http.Response {
	// TODO: Marshal a struct instead of using strings
	// <GhettoSanitize>
	project = strings.Replace(project, `"`, `'`, -1)
	ticketType = strings.Replace(ticketType, `"`, `'`, -1)
	subject = strings.Replace(subject, `"`, `'`, -1)
	description = strings.Replace(description, `"`, `'`, -1)
	assignee = strings.Replace(assignee, `"`, `'`, -1)
	assignee = strings.ToLower(assignee)
	// </GhettoSanitize>

	json := fmt.Sprintf(`{"project_id": "%v", "issue": {"tracker_id": %v, "subject": "%v", "description": "%v", "assigned_to_id": %v}}`,
		project, trackerID[ticketType], subject, description, assigneeToID[assignee])
	reader := strings.NewReader(json)

	client := &http.Client{}
	req, err := http.NewRequest("POST", REDMINE_JSON_URL, reader)
	req.Header.Add("X-Redmine-API-Key", BOT_REDMINE_KEY)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("client.Do error: %v\n", err)
		return nil
	}
	// fmt.Printf("%v\n", resp)
    // body, err := ioutil.ReadAll(req.Body)
	// if err != nil {
	// 	log.Printf("ReadAll Error: %v\n", err)
	// 	return nil
	// }
    // defer req.Body.Close()
    return resp
}
