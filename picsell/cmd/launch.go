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
	"encoding/json"
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/jweslley/localtunnel"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// launchCmd represents the launch command
var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cleanup()
			os.Exit(1)
		}()
		if len(args) < 1 {
			color.Red("Please specify the action to init ( picsell launch sweep ) \n")
			return
		}
		if args[0] == "sweep" {
			if len(args) < 2 {
				color.Red("Please provide the sweep ID ( picsell launch sweep SWEEP_ID ) \n")
				return
			}

			host_id := getConfigHost()
			url := URL + "run/cli/" + args[1] + "/next_run/" + host_id
			token := getAPIToken()
			var bearer = "Token " + token

			req, err := http.NewRequest("GET", url, nil)
			req.Header.Add("Authorization", bearer)

			client := &http.Client{}
			resp, err := client.Do(req)

			if err != nil {
				log.Fatal(err)
			}

			defer resp.Body.Close()

			var res Run
			json.NewDecoder(resp.Body).Decode(&res)
			var envs []string
			for i := range res.Env {
				envs = append(envs, res.Env[i].Name+"="+res.Env[i].Value)
			}

			fmt.Printf("Starting container %v to launch run %v %v \n", res.DockerImage, res.Experiment.Name, emoji.ManTechnologist)
			container_id := RunContainer(res.DockerImage, envs)

			fmt.Printf("%v started\n\nSee logs with ( docker logs %v ) in an other terminal %v \n\n", res.DockerImage, container_id, emoji.Laptop)
			var port = 8080
			subdomain := getConfigHost()

			var tunnel = localtunnel.NewLocalTunnel(port)
			var errs = tunnel.OpenAs(subdomain)
			if errs != nil {
				log.Fatal(err)
				return
			}

			fmt.Printf("Training %v launched, you can visualize your performance metrics on %v https://beta.picsellia.com %v  \n\n", res.Experiment.Name, emoji.Avocado, emoji.Avocado)
			color.Green("%v Do not kill this terminal, you won't be able to perform automatical job call %v", emoji.Warning, emoji.Warning)
			r := gin.Default()
			r.GET("/next_run", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "ip",
				})
			})
			r.Run()
		}
	},
}

func cleanup() {

	fmt.Println("Cleanup")
}
func init() {
	rootCmd.AddCommand(launchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// launchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// launchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
