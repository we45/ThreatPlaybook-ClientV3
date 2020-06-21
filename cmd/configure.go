/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

type ConfigObj struct {
	host     string
	port     int32
	email    string
	password string
	token    string
}

//Note, when you want to export a struct and use it to write, you need to capitalize the attributes
type JSONConfig struct {
	Host  string `json:"host"`
	Port  int32  `json:"port"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type LoginResponse struct {
	Success bool `json:"success"`
	Error   bool `json:"error"`
	Message bool `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configures the client to speak with the ThreatPlaybook API Server",
	Long:  `Configures the client to speak with the ThreatPlaybook API Server`,
	Run: func(cmd *cobra.Command, args []string) {
		password, _ := cmd.Flags().GetString("password")
		host, _ := cmd.Flags().GetString("url")
		port, _ := cmd.Flags().GetInt32("port")
		email, _ := cmd.Flags().GetString("email")

		loadConfiguration(ConfigObj{host: host, port: port, email: email, password: password})
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.Flags().StringP("url", "u", "localhost", "A Host FQDN or IPAddress without the scheme")
	configureCmd.Flags().Int32P("port", "p", 80, "Port to connect to server on. Default is 80")
	configureCmd.Flags().StringP("email", "e", "admin@admin.com", "Email used to authenticate to ThreatPlaybook")
	configureCmd.Flags().String("password", "", "Password to authenticate to ThreatPlaybook")
	configureCmd.MarkFlagRequired("host")
	configureCmd.MarkFlagRequired("port")
	configureCmd.MarkFlagRequired("email")
}

func getPasswordFromStdin() string {
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return ""
	} else {
		return strings.TrimSpace(string(bytePassword))
	}
}

func createOrUpdateCredFile(config ConfigObj) {
	jsonConfig := JSONConfig{
		Host:  config.host,
		Port:  config.port,
		Token: config.token,
		Email: config.email,
	}
	file, err := json.MarshalIndent(&jsonConfig, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	_ = ioutil.WriteFile(".cred", file, 0644)
}

func loadConfiguration(config ConfigObj) {
	var finalPassword string
	if config.password == "" {
		finalPassword = getPasswordFromStdin()
		if finalPassword == "" {
			log.Fatal("[-] There's no valid password entered")
		}
	} else {
		finalPassword = config.password
	}
	requestBody, json_err := json.Marshal(map[string]string{
		"email":    config.email,
		"password": finalPassword,
	})

	if json_err != nil {
		log.Fatal("[-] Unable to generate JSON request")
	}

	url := fmt.Sprintf("http://%s:%d/api/login", config.host, config.port)
	resp, httpError := http.Post(url, "application/json", bytes.NewBuffer(requestBody))

	if httpError != nil {
		log.Fatal(httpError)
	}
	if resp.StatusCode != 200 {
		fmt.Println(*resp)
		log.Fatal("[-] Unable to authenticate")
	}

	defer resp.Body.Close()
	var loginResponse LoginResponse

	json.NewDecoder(resp.Body).Decode(&loginResponse)
	config.token = loginResponse.Data.Token

	createOrUpdateCredFile(config)

	fmt.Println("[+] Successfully configured and logged in")

}
