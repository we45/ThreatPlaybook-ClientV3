package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type SecurityTestCase struct {
	Name      string `yaml:"name"`
	Test      string `yaml:"test"`
	Test_type string `yaml:"type"`
	Tools     string `yaml:"tools"`
}

type ScenarioReference struct {
	Name     string `yaml:"name"`
	Severity int32  `yaml:"severity"`
}

type ThreatScenario struct {
	Name         string             `yaml:"name"`
	ScenarioType string             `yaml:"type"`
	Description  string             `yaml:"description"`
	Vul_name     string             `yaml:"vul_name"`
	Severity     int32              `yaml:"severity"`
	Cwe          int32              `yaml:"cwe"`
	Reference    ScenarioReference  `yaml:"reference"`
	Test_cases   []SecurityTestCase `yaml:"test-cases"`
}

type AbuserStory struct {
	Name             string           `yaml:"name"`
	Description      string           `yaml:"description"`
	Threat_scenarios []ThreatScenario `yaml:"threat_scenarios"`
}

type Feature struct {
	ObjectType  string        `yaml:"objectType"`
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Abuse_cases []AbuserStory `yaml:"abuse_cases"`
}

type FeatureRequest struct {
	Name        string `json:"short_name"`
	Description string `json:"description"`
	Project     string `json:"project"`
}

type AbuseRequest struct {
	Name            string                  `json:"short_name"`
	Description     string                  `json:"description"`
	UserStory       string                  `json:"feature"`
	ThreatScenarios []ThreatScenarioRequest `json:"scenarios"`
}

type ThreatScenarioRequest struct {
	Name         string            `json:"name"`
	Feature      string            `json:"feature"`
	AbuserStory  string            `json:"abuser_story"`
	Description  string            `json:"description"`
	ScenarioType string            `json:"type"`
	RepoName     string            `json:"repo_name"`
	VulName      string            `json:"vul_name"`
	Cwe          int32             `json:"cwe"`
	Severity     int32             `json:"severity"`
	TestCases    []TestCaseRequest `json:"test_cases"`
	Categories   []string          `json:"categories"`
}

type TestCaseRequest struct {
	Name           string   `json:"name"`
	TestCase       string   `json:"test_case"`
	ThreatScenario string   `json:"threat_scenario"`
	Tools          []string `json:"tools"`
	TestType       string   `json:"test_type"`
}

type AbuseCaseParams struct {
	UseCase     string
	AbuserStory string
	Mode        string
	configValue ConfigObj
}

type AbuserStoryGetResponse struct {
	Data []struct {
		ShortName   string `json:"short_name"`
		Description string `json:"description"`
	} `json:"data"`
}

type ThreatScenarioGetResponse struct {
	Data []ThreatScenarioRequest `json:"data"`
}

type TestCaseGetResponse struct {
	Data []TestCaseRequest `json:"data"`
}

type CascadingFeatureObject struct {
	Name        string         `json:"short_name"`
	Description string         `json:"description"`
	AbuseCases  []AbuseRequest `json:"abuse_cases"`
}

func IsConfigured() bool {
	if _, err := os.Stat(".cred"); os.IsNotExist(err) {
		return false
	}
	return true
}

func GetJsonConfiguration() ConfigObj {
	jsonFile, err := os.Open(".cred")
	if err != nil {
		log.Fatal("Unable to open the credentials file.")
	}
	byteJson, _ := ioutil.ReadAll(jsonFile)
	var jsonConfig JSONConfig
	json.Unmarshal(byteJson, &jsonConfig)
	return ConfigObj{
		host:  jsonConfig.Host,
		port:  jsonConfig.Port,
		email: jsonConfig.Email,
		token: jsonConfig.Token,
	}
}

func ChangeTokenInCred(token string) {
	jsonFile, err := os.Open(".cred")
	if err != nil {
		log.Fatal("Unable to open the credentials file.")
	}
	byteJson, _ := ioutil.ReadAll(jsonFile)
	defer jsonFile.Close()
	var jsonConfig JSONConfig
	json.Unmarshal(byteJson, &jsonConfig)

	newConfig := JSONConfig{
		Host:  jsonConfig.Host,
		Port:  jsonConfig.Port,
		Email: jsonConfig.Email,
		Token: token,
	}
	file, err := json.MarshalIndent(&newConfig, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	_ = ioutil.WriteFile(".cred", file, 0644)
}

//MakeRequest - Generic Function to create an HTTP request
func MakeRequest(url string, method string, configValue ConfigObj, request []byte) *http.Response {
	client := &http.Client{}
	getReq, getErr := http.NewRequest(method, url, bytes.NewBuffer(request))
	getReq.Header.Add("Content-Type", "application/json")
	getReq.Header.Add("Authorization", configValue.token)
	if getErr != nil {
		log.Fatal("Unable to generate Get Project HTTP request")
	}

	getResp, respErr := client.Do(getReq)
	if respErr != nil {
		log.Fatal("Unable to make Get Project HTTP Request")
	}

	return getResp
}

//GetAbuserStory - Generic Function to get abuser stories for a user story or a specific abuser story
func GetAbuserStory(params AbuseCaseParams) AbuserStoryGetResponse {
	url := fmt.Sprintf("http://%s:%d/abuses/read", params.configValue.host, params.configValue.port)
	jsonPayload := map[string]string{}
	if params.UseCase != "" {
		jsonPayload["user_story"] = params.UseCase
	}

	if params.AbuserStory != "" {
		jsonPayload["short_name"] = params.AbuserStory
	}

	abRequest, abError := json.Marshal(jsonPayload)
	if abError != nil {
		log.Fatal("Unable to create Abuser Story JSON for Request")
	}

	getResp := MakeRequest(url, "POST", params.configValue, abRequest)

	defer getResp.Body.Close()

	body, err := ioutil.ReadAll(getResp.Body)
	if err != nil {
		panic(err.Error())
	}
	if getResp.StatusCode != 200 {
		fmt.Println("[ERROR]", body)
	}

	var allAbuses AbuserStoryGetResponse
	json.Unmarshal(body, &allAbuses)

	return allAbuses

}

//GetThreatScenario - Generic function to get ThreatScenarios for an Abuser Story or a Specific Threat Scenario
func GetThreatScenario(abuse_case string, config ConfigObj) ThreatScenarioGetResponse {
	url := fmt.Sprintf("http://%s:%d/scenarios/read", config.host, config.port)
	jsonPayload := map[string]string{
		"abuser_story": abuse_case,
	}
	abRequest, abError := json.Marshal(jsonPayload)
	if abError != nil {
		log.Fatal("Unable to create Threat Scenario JSON for Request")
	}

	getResp := MakeRequest(url, "POST", config, abRequest)
	defer getResp.Body.Close()

	body, err := ioutil.ReadAll(getResp.Body)
	if err != nil {
		panic(err.Error())
	}
	if getResp.StatusCode != 200 {
		fmt.Println("[ERROR]", body)
	}

	var allThreats ThreatScenarioGetResponse
	json.Unmarshal(body, &allThreats)
	return allThreats

}

//GetTestCase - Generic Function to get TestCases for a given scenario or a specific Test Case
func GetTestCase(scenario string, config ConfigObj) TestCaseGetResponse {
	url := fmt.Sprintf("http://%s:%d/test/read", config.host, config.port)
	jsonPayload := map[string]string{
		"scenario": scenario,
	}
	abRequest, abError := json.Marshal(jsonPayload)
	if abError != nil {
		log.Fatal("Unable to create Test Case JSON for Request")
	}

	getResp := MakeRequest(url, "POST", config, abRequest)
	defer getResp.Body.Close()

	body, err := ioutil.ReadAll(getResp.Body)
	if err != nil {
		panic(err.Error())
	}
	if getResp.StatusCode != 200 {
		fmt.Println("[ERROR]", body)
	}

	var allTests TestCaseGetResponse
	json.Unmarshal(body, &allTests)
	return allTests
}
