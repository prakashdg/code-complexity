package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	gitlabURL      = "https://gitlab.eng.vmware.com"
	projectId      = "13032"
	con_legend =`
| Score  | Label  | Description |
| ------ | ------ | ----------- |
| 0-9    | Easy   | Easily maintained code | 
| 10-19  | Little effort   | Maintained with little effort |
| 20-29  | Considerable effort | Maintained with considerable effort |
| 30-39  | Difficult   | Difficult to maintain code |
| 40-49  | Very difficult  | Very difficult to maintain code |
| 50-99  | Hard  | Hard to maintain code |
| 100 - 199  | Extremely hard  | Extremely hard to maintain code |
| 200+   | Unmaintanable  | Unmaintanable code *** consider rewrite of code *** |
	`
)

func check_gitlab_connection() {
	
	// Authenticate to GitLab API using personal access token
	gitlabToken:=os.Getenv("GITLAB_TOKEN")
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v4/projects/%s", gitlabURL, projectId), nil)
	req.Header.Set("PRIVATE-TOKEN", gitlabToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error: %s", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read the response body to verify authentication was successful
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error: %s %s", err)
		os.Exit(1)
	}
}

func find_and_replace_old_table(description string, table_desc string) (string, string) {

	re:=regexp.MustCompile(`(?s)Complexity table is.*Complexity table updated`)
	replaced:=re.ReplaceAllString(description, "")

	// TO-DO: Remove this dirty workarround with proper formater 
	table_up:=regexp.MustCompile(`T-LPF`)
	table_desc=table_up.ReplaceAllString(table_desc, "TargetBranch Lines Per Function")
	table_up=regexp.MustCompile(`S-LPF`)
	table_desc=table_up.ReplaceAllString(table_desc, "SourceBranch Lines Per Function")
	table_up=regexp.MustCompile(`T-FS`)
	table_desc=table_up.ReplaceAllString(table_desc, "TargetBranch Function Score")
	table_up=regexp.MustCompile(`S-FS`)
	table_desc=table_up.ReplaceAllString(table_desc, "SourceBranch Function Score")
	table_up=regexp.MustCompile(`FILENAME`)
	table_desc=table_up.ReplaceAllString(table_desc, "FileName")
	table_up=regexp.MustCompile(`FUNCNAME`)
	table_desc=table_up.ReplaceAllString(table_desc, "FuncName")

	return replaced, table_desc
}


func update_mr(mr_id string, description string, old_desc string) {
	
	// Check gitlab connectivity
	gitlabToken:=os.Getenv("GITLAB_TOKEN")
	check_gitlab_connection()
	client := &http.Client{}
	// Create the request body
	old_desc,description=find_and_replace_old_table(old_desc, description)
	accr:="\n$`\\textcolor{red}{\\text{*existing\\_function\\_updated}}`$ \n \n $`\\textcolor{blue}{\\text{*new\\_function\\_added}}`$\n"
	table_legend:=fmt.Sprintf("##### Complexity table legend \n %s", con_legend)
	new_description := fmt.Sprintf("%s \n\n ##### Complexity table is \n %s \n %s \n%s \n Complexity table updated", old_desc, description, accr, table_legend)
	body := map[string]string{
		"description": new_description,
	}
	requestBody, _ := json.Marshal(body)

	// Create the request
	url := fmt.Sprintf("https://gitlab.eng.vmware.com/api/v4/projects/13032/merge_requests/%s", mr_id)
	request, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(requestBody))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", gitlabToken))

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error: Failed", err)
		return
	}
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		fmt.Println("Error:Last failed", response.Status)
		return
	}

	fmt.Println("Merge Request description updated successfully!!")
}

func get_mr_desc(mr_id string) string {
	
	// Set up the API request
	gitlabToken:=os.Getenv("GITLAB_TOKEN")
	url := fmt.Sprintf("https://gitlab.eng.vmware.com/api/v4/projects/13032/merge_requests/%s", mr_id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	// Send the API request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Parse the JSON response
	var data struct {
		Description string `json:"description"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}
	return data.Description
}
