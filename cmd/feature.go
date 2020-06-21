// /*
// Copyright © 2020 Abhay Bhargav

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
// */
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type ThreatScenarioCreate struct {
	Name      string             //Threat Scenario name
	Created   bool               //Whether it was successfully created or not
	TestCases []SecurityTestCase //List of test cases, if its an inline object
}

type TestCaseCreate struct {
	Name    string //test-case name
	Created bool   //whether the test case was successfully created or not
}

type AbuserStoryCreate struct {
	Name            string
	Created         bool
	Feature         string
	ThreatScenarios []ThreatScenario
}

// featureCmd represents the feature command
var featureCmd = &cobra.Command{
	Use:   "feature",
	Short: "Load a feature/user story from a YAML file",
	Long:  `Upload a user story with abuser stories, threat scenarios and test cases to ThreatPlaybook with a YAML file`,
	Run: func(cmd *cobra.Command, args []string) {
		yfile, _ := cmd.Flags().GetString("file")
		projectName, _ := cmd.Flags().GetString("project")

		var doesProjectExist bool
		doesProjectExist = getProject(projectName)
		if doesProjectExist == true {
			fmt.Printf("[+] Project '%s' found. Moving on...\n", projectName)
			processFeatures(yfile, projectName)

		} else {
			log.Fatal("Unable to find project. Exiting...")
		}

	},
}

func init() {
	applyCmd.AddCommand(featureCmd)
	featureCmd.Flags().StringP("file", "f", "", "absolute file path to feature YAML file")
	featureCmd.Flags().StringP("project", "p", "", "project name that you want to tag the feature to")
	featureCmd.MarkFlagRequired("file")
	featureCmd.MarkFlagRequired("project")
}

func getProject(project string) bool {
	configValue := GetJsonConfiguration()
	if (ConfigObj{}) == configValue {
		log.Fatal("Unable to fetch value from cred file")
	}
	url := fmt.Sprintf("http://%s:%d/api/project/read", configValue.host, configValue.port)

	getProject := ProjectRequest{
		Name: project,
	}

	request, jsonErr := json.Marshal(getProject)

	if jsonErr != nil {
		log.Fatal("Unable to generate Request body as JSON")
	}

	client := &http.Client{}
	getReq, getErr := http.NewRequest("POST", url, bytes.NewBuffer(request))
	getReq.Header.Add("Content-Type", "application/json")
	getReq.Header.Add("Authorization", configValue.token)
	if getErr != nil {
		log.Fatal("Unable to generate Get Project HTTP request")
	}

	getResp, respErr := client.Do(getReq)
	if respErr != nil {
		log.Fatal("Unable to make Get Project HTTP Request")
	}

	defer getResp.Body.Close()

	if getResp.StatusCode != 200 {
		fmt.Println(getResp.Body)
		log.Fatal("Unable to get project or does not exist")
	} else {
		return true
	}

	return false

}

func processFeatures(filename string, project string) {
	yamlFile, fileErr := ioutil.ReadFile(filename)
	if fileErr != nil {
		log.Fatal("Unable to read YAML Feature file. Exiting...")
	}

	var userStory Feature

	parseError := yaml.Unmarshal(yamlFile, &userStory)
	if parseError != nil {
		log.Fatal("Unable to parse YAML File and generate data structure")
	}
	isUserStoryCreated := makeUserStory(userStory, project)
	if isUserStoryCreated == true {
		fmt.Println("Created Feature")
	} else {
		log.Fatalf("Unable to create feature '%s'\n", userStory.Name)
	}

	abuserStoriesStatus := makeAbuserStoryAndEverythingElse(userStory)
	for _, asValues := range abuserStoriesStatus {
		if asValues.Created == false {
			fmt.Printf("[😞] Unable to create abuser story %s\n", asValues.Name)
		}

		threatScenarios := makeThreatScenarios(asValues)
		for _, tsValues := range threatScenarios {
			if tsValues.Created == false {
				fmt.Printf("[😞] Unable to create threat scenario %s\n", tsValues.Name)
			}

			if len(tsValues.TestCases) > 0 {
				testCases := makeTestCases(tsValues)
				for _, caseVal := range testCases {
					if caseVal.Created == false {
						fmt.Printf("[😞] Unable to create test case %s\n", caseVal.Name)
					}
				}
			}

		}
	}

}

func makeUserStory(yamlFeature Feature, project string) bool {
	configValue := GetJsonConfiguration()
	if (ConfigObj{}) == configValue {
		log.Fatal("Unable to fetch value from cred file")
	}
	url := fmt.Sprintf("http://%s:%d/api/feature/create", configValue.host, configValue.port)

	makeFeature := FeatureRequest{
		Name:        yamlFeature.Name,
		Description: yamlFeature.Description,
		Project:     project,
	}
	request, jsonErr := json.Marshal(makeFeature)

	if jsonErr != nil {
		log.Fatal("Unable to generate Request body as JSON")
	}

	client := &http.Client{}
	getReq, getErr := http.NewRequest("POST", url, bytes.NewBuffer(request))
	getReq.Header.Add("Content-Type", "application/json")
	getReq.Header.Add("Authorization", configValue.token)
	if getErr != nil {
		log.Fatal("Unable to generate Make Feature HTTP request")
	}

	getResp, respErr := client.Do(getReq)
	if respErr != nil {
		log.Fatal("Unable to make Feature HTTP Request")
	}

	defer getResp.Body.Close()

	if getResp.StatusCode != 200 {
		fmt.Println(getResp.Body)
		return false
	}
	return true
}

func makeAbuserStoryAndEverythingElse(yamlFeature Feature) []AbuserStoryCreate {
	configValue := GetJsonConfiguration()
	if (ConfigObj{}) == configValue {
		log.Fatal("Unable to fetch value from cred file")
	}
	url := fmt.Sprintf("http://%s:%d/api/abuse-case/create", configValue.host, configValue.port)

	if len(yamlFeature.Abuse_cases) == 0 {
		log.Fatal("There are no abuse cases to be loaded. Exiting...")
	}

	var abuseStatus []AbuserStoryCreate

	allAbuses := yamlFeature.Abuse_cases

	//Abuse Case creation starts here
	for _, single := range allAbuses {
		makeAbuse := AbuseRequest{
			Name:        single.Name,
			Description: single.Description,
			UserStory:   yamlFeature.Name,
		}
		request, jsonErr := json.Marshal(makeAbuse)

		if jsonErr != nil {
			log.Fatal("Unable to generate Request body as JSON")
		}

		client := &http.Client{}
		getReq, getErr := http.NewRequest("POST", url, bytes.NewBuffer(request))
		getReq.Header.Add("Content-Type", "application/json")
		getReq.Header.Add("Authorization", configValue.token)
		if getErr != nil {
			log.Fatal("Unable to generate Make Abuser Story HTTP request")
		}

		getResp, respErr := client.Do(getReq)
		if respErr != nil {
			log.Fatal("Unable to make Abuser Story HTTP Request")
		}

		defer getResp.Body.Close()

		if getResp.StatusCode != 200 {
			fmt.Println(getResp.Body)
			errorAbuse := AbuserStoryCreate{
				Name:    single.Name,
				Created: false,
				Feature: yamlFeature.Name,
			}
			abuseStatus = append(abuseStatus, errorAbuse)
		}

		successAbuse := AbuserStoryCreate{
			Name:            single.Name,
			Created:         true,
			Feature:         yamlFeature.Name,
			ThreatScenarios: single.Threat_scenarios,
		}
		abuseStatus = append(abuseStatus, successAbuse)

		fmt.Printf("[+] Successfully created Abuser story '%s'\n", single.Name)
	}

	return abuseStatus

}

func makeThreatScenarios(singleAbuse AbuserStoryCreate) []ThreatScenarioCreate {
	configValue := GetJsonConfiguration()
	if (ConfigObj{}) == configValue {
		log.Fatal("Unable to fetch value from cred file")
	}
	repoURL := fmt.Sprintf("http://%s:%d/api/scenario/repo/create", configValue.host, configValue.port)
	inlineURL := fmt.Sprintf("http://%s:%d/api/scenario/create", configValue.host, configValue.port)

	if len(singleAbuse.ThreatScenarios) == 0 {
		log.Fatal("There are no threat scenarios to be loaded. Exiting...")
	}
	var url string
	var scenarioStatus []ThreatScenarioCreate

	allScenarios := singleAbuse.ThreatScenarios

	for _, single := range allScenarios {

		tsRequest := ThreatScenarioRequest{
			Name:        single.Name,
			Description: single.Description,
			Feature:     singleAbuse.Feature,
			AbuserStory: singleAbuse.Name,
		}

		switch single.ScenarioType {
		case "repo":
			if single.Reference.Name != "" && single.Reference.Severity > 0 {
				tsRequest.ScenarioType = "repo"
				tsRequest.RepoName = single.Reference.Name
				tsRequest.Severity = single.Reference.Severity
			}
			url = repoURL
		case "inline":
			tsRequest.ScenarioType = "inline"
			tsRequest.Severity = single.Severity
			if single.Cwe != 0 {
				tsRequest.Cwe = single.Cwe
			}
			url = inlineURL
		default:
			fmt.Printf("Threat Scenario '%s' is not 'repo' or 'inline'. These are the only options. Ignoring...\n", single.Name)
		}

		scenarioRequest, jsonErr := json.Marshal(tsRequest)
		if jsonErr != nil {
			log.Fatalf("Unable to serialize JSON for Threat Scenario: %s\n", single.Name)
		}

		client := &http.Client{}

		getReq, getErr := http.NewRequest("POST", url, bytes.NewBuffer(scenarioRequest))
		getReq.Header.Add("Content-Type", "application/json")
		getReq.Header.Add("Authorization", configValue.token)
		if getErr != nil {
			log.Fatal("Unable to generate Make Threat Scenario HTTP request")
		}

		getResp, respErr := client.Do(getReq)
		if respErr != nil {
			log.Fatal("Unable to make Threat Scenario HTTP Request")
		}

		defer getResp.Body.Close()

		if getResp.StatusCode != 200 {
			singleStatus := ThreatScenarioCreate{
				Name:    single.Name,
				Created: false,
			}
			if single.ScenarioType == "inline" {
				singleStatus.TestCases = single.Test_cases
			}
			scenarioStatus = append(scenarioStatus, singleStatus)
			fmt.Printf("Unable to load Threat Scenario '%s'\n", single.Name)
		}

		successScenario := ThreatScenarioCreate{
			Name:    single.Name,
			Created: true,
		}
		if single.ScenarioType == "inline" {
			successScenario.TestCases = single.Test_cases
		}
		scenarioStatus = append(scenarioStatus, successScenario)
		fmt.Printf("[+] Successfully loaded Threat Scenario '%s'\n", single.Name)

	}

	return scenarioStatus

}

func makeTestCases(singleScenario ThreatScenarioCreate) []TestCaseCreate {
	configValue := GetJsonConfiguration()
	if (ConfigObj{}) == configValue {
		log.Fatal("Unable to fetch value from cred file")
	}
	url := fmt.Sprintf("http://%s:%d/api/test/create", configValue.host, configValue.port)

	var testCaseStatus []TestCaseCreate

	allCases := singleScenario.TestCases

	for _, single := range allCases {
		singleTC := TestCaseRequest{
			Name:           single.Name,
			TestCase:       single.Test,
			ThreatScenario: singleScenario.Name,
		}

		if single.Tools != "" {
			allTools := strings.Split(single.Tools, ",")
			singleTC.Tools = allTools
		}

		if single.Test_type != "" {
			singleTC.TestType = single.Test_type
		}

		singleCaseRequest, jsonErr := json.Marshal(singleTC)
		if jsonErr != nil {
			log.Fatalf("Unable to serialize JSON for Test Case: %s\n", single.Name)
		}

		client := &http.Client{}

		getReq, getErr := http.NewRequest("POST", url, bytes.NewBuffer(singleCaseRequest))
		getReq.Header.Add("Content-Type", "application/json")
		getReq.Header.Add("Authorization", configValue.token)
		if getErr != nil {
			log.Fatal("Unable to generate Make Test Case HTTP request")
		}

		getResp, respErr := client.Do(getReq)
		if respErr != nil {
			log.Fatal("Unable to make Test Case HTTP Request")
		}

		defer getResp.Body.Close()

		if getResp.StatusCode != 200 {
			singleStatus := TestCaseCreate{
				Name:    single.Name,
				Created: false,
			}

			testCaseStatus = append(testCaseStatus, singleStatus)
			fmt.Printf("Unable to load Test Case '%s'\n", single.Name)
		}

		successScenario := TestCaseCreate{
			Name:    single.Name,
			Created: true,
		}
		testCaseStatus = append(testCaseStatus, successScenario)
		fmt.Printf("[+] Successfully loaded Test Case '%s'\n", single.Name)

	}

	return testCaseStatus
}
