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
	var stream bool
	flag.StringVar(&apiKey, "apiKey", "", "github api key to use, can also be GITHUB_API_KEY environment variable")
	flag.StringVar(&name, "name", "", "name of the user/org to list")
	flag.BoolVar(&isOrg, "isOrg", false, "is this an org (opposed to a user)")
	flag.StringVar(&apiPoint, "apiPoint", "api.github.com", "api endpoint to use")
	flag.StringVar(&action, "action", "repos", "action to paginate on")
	flag.StringVar(&prop, "prop", "name", "property to produce")
	flag.BoolVar(&stream, "stream", false, "output as a stream of values, instead of an array")

	flag.Parse()

	var orgVusr string
	if isOrg {
		orgVusr = "orgs"
	} else {
		orgVusr = "users"
	}

	if apiKey == "" {
		apiKey = os.Getenv("GITHUB_API_KEY")
	}

	url := fmt.Sprintf("https://%s/%s/%s/%s", apiPoint, orgVusr, name, action)

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("Couldn't create request:", err)
	}

	var output []interface{}

	enc := json.NewEncoder(os.Stdout)
	for {
		if apiKey != "" {
			req.Header.Set("Authorization", fmt.Sprintf("token %s", apiKey))
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalln("Could not initiate request:", err)
		}
		if resp.StatusCode != 200 {
			log.Fatalln("Got bad status code:", resp.StatusCode)
		}
		links := webLinks.Parse(resp.Header.Get("Link"))
		linksMap := links.Map()

		decoder := json.NewDecoder(resp.Body)
		var values []map[string]interface{}
		decoder.Decode(&values)
		resp.Body.Close()
		for _, value := range values {
			if stream {
				err = enc.Encode(value[prop])
				if err != nil {
					log.Fatalln("Could not encode value:", err)
				}
			} else {
				output = append(output, value[prop])
			}
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

	if !stream {
		err = enc.Encode(output)
		if err != nil {
			log.Fatalln("Could not encode output:", err)
		}
	}

}
