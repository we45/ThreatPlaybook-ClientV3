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
	"io/ioutil"
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type GetUserStoryRequest struct {
	ShortName   string `json:"short_name"`
	Project     string `json:"project"`
	Description string `json:"description"`
}

type UserStoryResponse struct {
	Data struct {
		GetUserStoryRequest
	} `json:"data"`
}

// userStoryCmd represents the userStory command
var userStoryCmd = &cobra.Command{
	Use:   "userStory",
	Short: "Gets a user-story based on the short-name of that user-story",
	Long:  `Gets a user-story and outputs that user story in json. Allows you to select the user-story, abuser-story and scenarios, test-cases based on cascading command`,
	Run: func(cmd *cobra.Command, args []string) {
		shortName, _ := cmd.Flags().GetString("short-name")
		project, _ := cmd.Flags().GetString("project")
		cascade, _ := cmd.Flags().GetBool("cascade")
		format, _ := cmd.Flags().GetString("format")

		outputUserStory(shortName, project, format, cascade)

	},
}

func init() {
	getCmd.AddCommand(userStoryCmd)
	userStoryCmd.Flags().StringP("short-name", "n", "", "Short name for the user-story to query by. Mandatory")
	userStoryCmd.Flags().StringP("project", "p", "", "Project that this user story belongs to. Mandatory")
	userStoryCmd.Flags().BoolP("cascade", "r", false, "Gets you the user story, its associated abuser stories and threat scenarios. Will output as JSON always.")
	userStoryCmd.Flags().StringP("format", "f", "stdout", "Default output format for the dataset. Default is stdout tables")

	userStoryCmd.MarkFlagRequired("short-name")
	userStoryCmd.MarkFlagRequired("project")

}

func cascadeFeatureToScenario(configValue ConfigObj, name string, project string) {
	url := fmt.Sprintf("http://%s:%d/feature/read", configValue.host, configValue.port)
	getUserStory := GetUserStoryRequest{
		ShortName: name,
		Project:   project,
	}

	guRequest, jsonErr := json.Marshal(getUserStory)
	if jsonErr != nil {
		log.Fatal("Unable to serialize JSON to query user stories")
	}

	getResp := MakeRequest(url, "POST", configValue, guRequest)

	defer getResp.Body.Close()

	if getResp.StatusCode != 200 {
		fmt.Println(getResp.Body)
		log.Fatal("Unable to get project or does not exist")
	} else {
		body, err := ioutil.ReadAll(getResp.Body)
		if err != nil {
			panic(err.Error())
		}
		var respData UserStoryResponse
		json.Unmarshal(body, &respData)

	}

}

func outputUserStory(name string, project string, format string, cascade bool) {
	configValue := GetJsonConfiguration()
	if (ConfigObj{}) == configValue {
		log.Fatal("Unable to fetch value from cred file. Have you configured and logged in yet?")
	}
	url := fmt.Sprintf("http://%s:%d/feature/read", configValue.host, configValue.port)

	getUserStory := GetUserStoryRequest{
		ShortName: name,
		Project:   project,
	}

	guRequest, jsonErr := json.Marshal(getUserStory)
	if jsonErr != nil {
		log.Fatal("Unable to serialize JSON to query user stories")
	}

	getResp := MakeRequest(url, "POST", configValue, guRequest)

	defer getResp.Body.Close()
	var respData UserStoryResponse
	if getResp.StatusCode != 200 {
		fmt.Println(getResp.Body)
		log.Fatal("Unable to get project or does not exist")
	}
	body, err := ioutil.ReadAll(getResp.Body)
	if err != nil {
		panic(err.Error())
	}
	if format == "json" {
		fmt.Println(string(body))
	} else {
		json.Unmarshal(body, &respData)
	}

	if cascade == false {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Description", "Project"})
		table.Append([]string{
			respData.Data.ShortName,
			respData.Data.Description,
			project,
		})
		table.Render()
	} else {
		var cascadingFeature CascadingFeatureObject
		cascadingFeature.Name = respData.Data.ShortName
		cascadingFeature.Description = respData.Data.Description
		allAbuses := GetAbuserStory(AbuseCaseParams{
			UseCase:     name,
			configValue: configValue,
			Mode:        "cascade",
		})
		for _, story := range allAbuses.Data {
			newAbuse := AbuseRequest{
				Name:        story.ShortName,
				Description: story.Description,
			}
			allThreats := GetThreatScenario(newAbuse.Name, configValue)

			for _, scenario := range allThreats.Data {
				newScenario := ThreatScenarioRequest{
					Name:       scenario.Name,
					Cwe:        scenario.Cwe,
					Categories: scenario.Categories,
					VulName:    scenario.VulName,
				}
				allTests := GetTestCase(newScenario.Name, configValue)

				for _, test := range allTests.Data {
					if test.Name != "" {
						scenario.TestCases = append(scenario.TestCases, test)
					}
				}
				newAbuse.ThreatScenarios = append(newAbuse.ThreatScenarios, scenario)
			}
			cascadingFeature.AbuseCases = append(cascadingFeature.AbuseCases, newAbuse)
		}

		strJ, _ := json.Marshal(cascadingFeature)
		fmt.Println(string(strJ))
	}

}
