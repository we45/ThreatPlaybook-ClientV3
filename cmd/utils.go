package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
	Name        string `json:"short_name"`
	Description string `json:"description"`
	UserStory   string `json:"feature"`
}

type ThreatScenarioRequest struct {
	Name         string `json:"name"`
	Feature      string `json:"feature"`
	AbuserStory  string `json:"abuser_story"`
	Description  string `json:"description"`
	ScenarioType string `json:"type"`
	RepoName     string `json:"repo_name"`
	VulName      string `json:"vul_name"`
	Cwe          int32  `json:"cwe"`
	Severity     int32  `json:"severity"`
}

type TestCaseRequest struct {
	Name           string   `json:"name"`
	TestCase       string   `json:"test_case"`
	ThreatScenario string   `json:"threat_scenario"`
	Tools          []string `json:"tools"`
	TestType       string   `json:"test_type"`
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
