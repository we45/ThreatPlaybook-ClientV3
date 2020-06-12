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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

type ProjectRequest struct {
	Name string `json:"name"`
}

type ProjectResponse struct {
	Success bool `json:"success"`
	Error   bool `json:"error"`
	Data    struct {
		Name string `json:"name"`
	} `json:"data"`
}

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Creates a Project in ThreatPlaybook with the API",
	Long:  `Creates a project with a Project name denoted by the '-n' or '--name' parameter`,
	Run: func(cmd *cobra.Command, args []string) {

		name, _ := cmd.Flags().GetString("name")

		if name != "" {
			createProjectRequest(name)
		} else {
			log.Fatal("No name specified")
		}

	},
}

func init() {
	applyCmd.AddCommand(projectCmd)
	projectCmd.Flags().StringP("name", "n", "", "name of the project to be created")
	projectCmd.MarkFlagRequired("name")

}

func createProjectRequest(project string) {
	configValue := GetJsonConfiguration()
	if (ConfigObj{}) == configValue {
		log.Fatal("Unable to fetch value from cred file")
	}
	url := fmt.Sprintf("http://%s:%d/project/create", configValue.host, configValue.port)

	createProject := ProjectRequest{
		Name: project,
	}

	request, jsonErr := json.Marshal(createProject)

	if jsonErr != nil {
		log.Fatal("Unable to generate Request body as JSON")
	}

	client := &http.Client{}
	getReq, getErr := http.NewRequest("POST", url, bytes.NewBuffer(request))
	getReq.Header.Add("Content-Type", "application/json")
	getReq.Header.Add("Authorization", configValue.token)
	if getErr != nil {
		log.Fatal("Unable to generate Project HTTP request")
	}

	getResp, respErr := client.Do(getReq)
	if respErr != nil {
		log.Fatal("Unable to make Project HTTP Request")
	}

	defer getResp.Body.Close()

	if getResp.StatusCode != 200 {
		fmt.Println(getResp.Body)
		log.Fatal("Unable to create project")
	}

	var projectResp ProjectResponse
	byteBody, berr := ioutil.ReadAll(getResp.Body)
	if berr != nil {
		log.Fatal("Unable to parse Project response body. But project is probably created")
	}
	projErr := json.Unmarshal(byteBody, &projectResp)

	if projErr != nil {
		log.Fatal("Unable to create project. Response doesnt match")
	}

	fmt.Printf("[+] Project '%s' successfully created\n", projectResp.Data.Name)
}
