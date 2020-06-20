/*
Copyright © 2020 Abhay Bhargav

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		delObject, _ := cmd.Flags().GetString("object")
		objName, _ := cmd.Flags().GetString("name")
		delParent, _ := cmd.Flags().GetString("parent")

		configValue := GetJsonConfiguration()
		if (ConfigObj{}) == configValue {
			log.Fatal("Unable to fetch value from cred file. Have you configured and logged in yet?")
		}

		switch delObject {
		case "feature":
			if delParent == "" {
				log.Fatal("Please enter a project name to delete the feature from")
			}
			deleteFeature(objName, delParent, configValue)
		case "abuser-story":
			if delParent == "" {
				log.Fatal("Please enter a feature name to delete the abuser-story from")
			}
			deleteAbuserStory(objName, delParent, configValue)
		case "scenario":
			if delParent == "" {
				log.Fatal("Please enter a abuser story short-name to delete the threat-scenario from")
			}
			deleteThreatScenario(objName, delParent, configValue)
		case "test":
			if delParent == "" {
				log.Fatal("Please enter a scenario to refer to, for deleting the test-case")
			}
			deleteTestCase(objName, delParent, configValue)
		case "project":
			deleteProject(objName, configValue)
		default:
			fmt.Println("Didnt understand any of these ob")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringP("object", "o", "", "Object like `feature`, `abuser-story`, `scenario`, `test` need to be mentioned here")
	deleteCmd.Flags().StringP("name", "n", "", "Name of the object to be deleted")
	deleteCmd.Flags().StringP("parent", "p", "", "Parent object Name reference name of the object that you are trying to delete. This is not required only for Project deletes. Required for everything else")
	deleteCmd.MarkFlagRequired("object")
	deleteCmd.MarkFlagRequired("name")
}

func deleteFeature(name string, project string, config ConfigObj) {

	prompt := promptui.Prompt{
		Label:     "Are you sure you want to delete?",
		IsConfirm: true,
	}
	result, err := prompt.Run()

	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	if result == "yes" || result == "y" || result == "YES" {
		url := fmt.Sprintf("http://%s:%d/delete/feature", config.host, config.port)

		jsonVal := map[string]string{
			"name":    name,
			"project": project,
		}
		delRequest, jsonErr := json.Marshal(jsonVal)
		if jsonErr != nil {
			log.Fatal("Unable to serialize JSON to query user stories")
		}

		getResp := MakeRequest(url, "POST", config, delRequest)

		defer getResp.Body.Close()

		if getResp.StatusCode != 200 {
			fmt.Println(getResp.Body)
			log.Fatal("Unable to get project or does not exist")
		}
		fmt.Printf("Feature: '%s' successfully deleted", name)
	}
}

func deleteAbuserStory(name string, feature string, config ConfigObj) {

	prompt := promptui.Prompt{
		Label:     "Are you sure you want to delete?",
		IsConfirm: true,
	}
	result, err := prompt.Run()

	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	if result == "yes" || result == "y" || result == "YES" {
		url := fmt.Sprintf("http://%s:%d/delete/abuser-story", config.host, config.port)

		jsonVal := map[string]string{
			"name":    name,
			"feature": feature,
		}
		delRequest, jsonErr := json.Marshal(jsonVal)
		if jsonErr != nil {
			log.Fatal("Unable to serialize JSON to delete abuser stories")
		}

		getResp := MakeRequest(url, "POST", config, delRequest)

		defer getResp.Body.Close()

		if getResp.StatusCode != 200 {
			fmt.Println(getResp.Body)
			log.Fatalf("Unable to delete Abuser Story: %s\n", name)
		}
		fmt.Printf("Abuser Story '%s' successfully deleted", name)
	}
}

func deleteThreatScenario(name string, abuserStory string, config ConfigObj) {

	prompt := promptui.Prompt{
		Label:     "Are you sure you want to delete?",
		IsConfirm: true,
	}
	result, err := prompt.Run()

	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	if result == "yes" || result == "y" || result == "YES" {
		url := fmt.Sprintf("http://%s:%d/delete/scenario", config.host, config.port)

		jsonVal := map[string]string{
			"name":         name,
			"abuser_story": abuserStory,
		}
		delRequest, jsonErr := json.Marshal(jsonVal)
		if jsonErr != nil {
			log.Fatal("Unable to serialize JSON to delete Threat Scenarios")
		}

		getResp := MakeRequest(url, "POST", config, delRequest)

		defer getResp.Body.Close()

		if getResp.StatusCode != 200 {
			fmt.Println(getResp.Body)
			log.Fatalf("Unable to delete Threat Scenario: %s\n", name)
		}
		fmt.Printf("Threat Scenario '%s' successfully deleted", name)
	}
}

func deleteTestCase(name string, scenario string, config ConfigObj) {

	prompt := promptui.Prompt{
		Label:     "Are you sure you want to delete?",
		IsConfirm: true,
	}
	result, err := prompt.Run()

	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	if result == "yes" || result == "y" || result == "YES" {
		url := fmt.Sprintf("http://%s:%d/delete/test", config.host, config.port)

		jsonVal := map[string]string{
			"name":     name,
			"scenario": scenario,
		}
		delRequest, jsonErr := json.Marshal(jsonVal)
		if jsonErr != nil {
			log.Fatal("Unable to serialize JSON to delete Threat Scenarios")
		}

		getResp := MakeRequest(url, "POST", config, delRequest)

		defer getResp.Body.Close()

		if getResp.StatusCode != 200 {
			fmt.Println(getResp.Body)
			log.Fatalf("Unable to delete Test Case: %s\n", name)
		}
		fmt.Printf("Test Case '%s' successfully deleted", name)
	}
}

func deleteProject(name string, config ConfigObj) {

	prompt := promptui.Prompt{
		Label:     "Are you sure you want to delete?",
		IsConfirm: true,
	}
	result, err := prompt.Run()

	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	if result == "yes" || result == "y" || result == "YES" {
		url := fmt.Sprintf("http://%s:%d/delete/project", config.host, config.port)

		jsonVal := map[string]string{
			"name": name,
		}
		delRequest, jsonErr := json.Marshal(jsonVal)
		if jsonErr != nil {
			log.Fatal("Unable to serialize JSON to delete Project")
		}

		getResp := MakeRequest(url, "POST", config, delRequest)

		defer getResp.Body.Close()

		if getResp.StatusCode != 200 {
			fmt.Println(getResp.Body)
			log.Fatalf("Unable to delete Project: %s\n", name)
		}
		fmt.Printf("Project '%s' successfully deleted", name)
	}
}
