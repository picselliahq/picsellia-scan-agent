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
	"log"
	"net/http"
	"os/exec"
	"os/user"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var URL = "https://beta.picsellia.com/sdk/v2/"

type Configuration struct {
	TunnelUrlID string
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init sweep YOUR_SWEEP_ID",
	Short: "Initialize connection to Picsellia server",
	Long:  `Use the init command to register this machine to Picsellia platform, This way Picsellia Oracle will be able to send jobs to it.`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			color.Red("Please specify the action to init ( picsell init sweep ID ) \n")
			return
		}
		user, _ := user.Current()
		if args[0] == "sweep" {
			if len(args) < 2 {
				color.Red("Please provide the sweep ID ( picsell init sweep SWEEP_ID ) \n")
				return
			}
			subdomain := getConfigHost()
			var tunnelUrl = "https://" + subdomain + ".loca.lt"

			values := map[string]string{"tunnelUrl": tunnelUrl, "username": user.Username, "id": subdomain}
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
				log.Fatalf("%v  Picsell platform not accessible, please try again later  %v \n", emoji.Warning, emoji.Warning)
				return
			}

			if resp.StatusCode == http.StatusNotFound {
				color.Red("The sweep you are trying to init does not exists")
				return
			}

			if resp.StatusCode == http.StatusNoContent {
				color.Green("No more jobs to run !")
				return
			}

			defer resp.Body.Close()

			var res DockerRun
			json.NewDecoder(resp.Body).Decode(&res)

			fmt.Printf("Pulling docker image %v %v\n", res.DockerImage, emoji.Whale)

			app := "docker"
			arg0 := "pull"
			arg1 := res.DockerImage

			shutdownCh := make(chan struct{})
			go indicator(shutdownCh)
			cmd := exec.Command(app, arg0, arg1)
			_, err = cmd.Output()

			close(shutdownCh)
			fmt.Printf("%v\n", emoji.CheckMark)

			if err != nil {
				color.Red("Picsell can't pull %v\nConsider running ( docker pull %v ) in another terminal", res.DockerImage, res.DockerImage)
				return
			}

			fmt.Printf("Docker image %v pulled %v\nYou can run (picsell launch sweep %v ) \n", res.DockerImage, emoji.ThumbsUp, args[1])
		} else {
			color.Red("Please run init sweep ID")
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
