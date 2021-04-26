/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"github.com/enescakir/emoji"
	//"github.com/jweslley/localtunnel"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os/exec"
	"os/user"
)

var URL = "http://localhost:8000/sdk/v2/"

type Configuration struct {
	TunnelUrlID string
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			fmt.Printf("Please specify the action to init ( picsell init sweep ) \n")
			return
		}
		user, _ := user.Current()
		if args[0] == "sweep" {
			if len(args) < 2 {
				fmt.Printf("Please provide the sweep ID ( picsell init sweep SWEEP_ID ) \n")
				return
			}
			subdomain := getConfigHost()
			var tunnelUrl = "https://" + subdomain + ".loca.lt"

			values := map[string]string{"tunnelUrl": tunnelUrl, "username": user.Username}
			json_data, err := json.Marshal(values)
			if err != nil {
				log.Fatal(err)
			}

			token := getAPIToken()
			var bearer = "Token " + token

			req, err := http.NewRequest("POST", URL+"sweep/"+args[1]+"/register", bytes.NewBuffer(json_data))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", bearer)

			client := &http.Client{}
			resp, err := client.Do(req)

			if err != nil {
				log.Fatal(err)
			}

			defer resp.Body.Close()

			var res DockerRun
			json.NewDecoder(resp.Body).Decode(&res)

			fmt.Printf("Pulling docker image %v %v \n \n", res.DockerImage, emoji.Whale)

			app := "docker"

			arg0 := "pull"
			arg1 := res.DockerImage

			cmd := exec.Command(app, arg0, arg1)
			stdout, err := cmd.Output()

			if err != nil {
				fmt.Println(err.Error())
				return
			}
			// Print the output
			fmt.Println(string(stdout))
			fmt.Printf("Docker image %v pulled %v \n You can run (picsell launch sweep %v ) \n", res.DockerImage, emoji.ThumbsUp, args[1])
		}

	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
