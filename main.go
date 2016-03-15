package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/conslo/webLinks"
)

func main() {
	var apiKey string
	var name string
	var isOrg bool
	var apiPoint string
	var action string
	var prop string
	flag.StringVar(&apiKey, "apiKey", "", "github api key to use")
	flag.StringVar(&name, "name", "", "name of the user/org to list")
	flag.BoolVar(&isOrg, "isOrg", false, "is this an org (opposed to a user)")
	flag.StringVar(&apiPoint, "apiPoint", "api.github.com", "api endpoint to use")
	flag.StringVar(&action, "action", "repos", "action to paginate on")
	flag.StringVar(&prop, "prop", "name", "property to produce")

	flag.Parse()

	var orgVusr string
	if isOrg {
		orgVusr = "orgs"
	} else {
		orgVusr = "users"
	}

	url := fmt.Sprintf("https://%s/%s/%s/%s", apiPoint, orgVusr, name, action)

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("Couldn't create request:", err)
	}

	var output []interface{}

	for {
		if apiKey != "" {
			req.Header.Set("Authorization", fmt.Sprintf("token %s", apiKey))
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalln("Could not initiate request:", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Fatalln("Got bad status code:", resp.StatusCode)
		}
		links := webLinks.Parse(resp.Header.Get("Link"))
		linksMap := links.Map()

		decoder := json.NewDecoder(resp.Body)
		var values []map[string]interface{}
		decoder.Decode(&values)
		for _, value := range values {
			output = append(output, value[prop])
		}
		next, ok := linksMap["next"]
		if !ok {
			break
		}
		req, err = http.NewRequest("GET", next.URI, nil)
		if err != nil {
			log.Fatalln("Could not make request from parsed url:", err)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(output)

}
