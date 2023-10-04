package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func getconfig() (string, string, string, string, string) {
	// Read the content of the aconfig.json file
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println("Error reading config.json:", err)
		return "", "", "", "", ""
	}

	// Parse the JSON data into the Config struct
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", "", "", "", ""
	}
	return config.Client_id, config.Redirect_uri, config.Grant_type, config.Scope, config.Medwerker_Email
}

func getsecrets() (string, string) {
	State := os.Getenv("SSO_STATE")
	Client_Secret := os.Getenv("SSO_CLIENT_SECRET")
	return State, Client_Secret
}

func containsSubstring(inputString, substring string) bool {
	return strings.Contains(inputString, substring)
}

type Config struct {
	Client_id          string `json:"client_id"`
	Redirect_uri       string `json:"redirect_uri"`
	Grant_type         string `json:"grant_type"`
	Grant_type_refresh string `json:"grant_type_refresh"`
	Scope              string `json:"scope"`
	Medwerker_Email    string `json:"medwerker_email"`
}

type microsoft_access struct {
	AccessToken string `json:"access_token"`
}

type microsoft_graph struct {
	Account []struct {
		AgeGroup    string `json:"ageGroup"`
		CountryCode string `json:"countryCode"`
		Email       string `json:"userPrincipalName"`
		UUID        string `json:"id"`
	} `json:"account"`
	Emails []struct {
		Email string `json:"address"`
	} `json:"emails"`
	Names []struct {
		First string `json:"first"`
		Last  string `json:"last"`
	}
}

func main() {
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		secrets_state, client_secret := getsecrets()
		if secrets_state == "" {
			http.Error(w, "Error reading secrets.json", http.StatusBadRequest)
			return
		}
		if code == "" {
			http.Error(w, "Error 400: No code provided.", http.StatusBadRequest)
			return
		} else if state == "" {
			http.Error(w, "Error 400: No key provided.", http.StatusBadRequest)
			return
		} else if state != secrets_state {
			http.Error(w, "Error 401: key not valid.", http.StatusUnauthorized)
			return
		} else {
			// Define the token endpoint URL
			tokenURL := "https://login.microsoftonline.com/common/oauth2/v2.0/token"

			// Prepare the form data
			formData := url.Values{}
			client_id, redirect_uri, grant_type, scope, medwerker_email := getconfig()
			if client_id == "" {
				http.Error(w, "Error reading config.json", http.StatusBadRequest)
				return
			}
			formData.Set("client_id", client_id)
			formData.Set("code", code)
			formData.Set("scope", scope)
			formData.Set("redirect_uri", redirect_uri)
			formData.Set("grant_type", grant_type)
			formData.Set("client_secret", client_secret)

			// Create a new POST request
			req, err := http.NewRequest("POST", tokenURL, strings.NewReader(formData.Encode()))
			if err != nil {
				fmt.Println("Error creating request:", err)
				return
			}

			// Set the request headers
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Send the request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error sending request:", err)
				http.Error(w, "Error sending request to Microsoft API", http.StatusGatewayTimeout)
				return
			}
			defer resp.Body.Close()

			// Check the response status code
			if resp.StatusCode != http.StatusOK {
				fmt.Println("Unexpected response status:", resp.Status)
				http.Error(w, "Unexpected response status from Microsoft API \nCheck if Access-Token(code) is valid", http.StatusBadRequest)
				return
			}

			// Parse the response body as JSON
			var microsoft_access microsoft_access
			err = json.NewDecoder(resp.Body).Decode(&microsoft_access)
			if err != nil {
				fmt.Println("Error decoding JSON response:", err)
				http.Error(w, "Error decoding JSON response from Microsoft API", http.StatusNotAcceptable)
				return
			}

			AgeGroup, CountryCode, UUID, eMail, FirstName, LastName := Graph_Microsoft(microsoft_access.AccessToken)
			Name := FirstName + " " + LastName
			Employee := containsSubstring(eMail, medwerker_email)

			// Create the JSON response with the desired format
			jsonResponse := fmt.Sprintf(`{
	"AgeGroup": "%s",
	"CountryCode": "%s",
	"UUID": "%s",
	"eMail": "%s",
	"Employee": "%v",
	"Name": "%s",
	"FirstName": "%s",
	"LastName": "%s"
}`, AgeGroup, CountryCode, UUID, eMail, Employee, Name, FirstName, LastName)

			// Set the Content-Type header and write the JSON response
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, jsonResponse)
			return
		}
	})
	log.Fatal(http.ListenAndServe(":80", nil))
}

func Graph_Microsoft(token string) (AgeGroup string, CountryCode string, UUID string, eMail string, FirstName string, LastName string) {
	url := "https://graph.microsoft.com/beta/me/profile"

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the content-type and accept headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Request successful!")
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}

		// Parse the JSON response
		var Microsoft_Graph microsoft_graph
		err = json.Unmarshal(respBody, &Microsoft_Graph)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}

		// get the info from the json response
		AgeGroup := Microsoft_Graph.Account[0].AgeGroup
		CountryCode := Microsoft_Graph.Account[0].CountryCode
		UUID := Microsoft_Graph.Account[0].UUID
		eMail := Microsoft_Graph.Emails[0].Email
		FirstName := Microsoft_Graph.Names[0].First
		LastName := Microsoft_Graph.Names[0].Last
		return AgeGroup, CountryCode, UUID, eMail, FirstName, LastName
	} else {
		fmt.Println("Request failed with status code:", resp.StatusCode)

	}
	return "Request failed", "", "", "", "", ""
}
