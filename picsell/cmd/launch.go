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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/jweslley/localtunnel"
	"github.com/spf13/cobra"
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
			fmt.Printf("%v Exiting run, the machine will be removed from available host for your sweep %v \n", emoji.Avocado, emoji.Avocado)
			unregisterHost(args[1])
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

			fmt.Printf("Starting container %v to launch run %v %v \n", res.DockerImage, res.Name, emoji.ManTechnologist)
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

			fmt.Printf("Training %v launched, you can visualize your performance metrics on %v https://beta.picsellia.com %v  \n\n", res.Name, emoji.Avocado, emoji.Avocado)
			color.Green("%v  Do not kill this terminal, you won't be able to perform automatical job call  %v\n", emoji.Warning, emoji.Warning)
			r := gin.Default()

			srv := &http.Server{
				Addr:    ":8080",
				Handler: r,
			}

			r.POST("/next_run", func(c *gin.Context) {
				var run Run
				if err := c.BindJSON(&run); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "bad request format",
					})
					return
				}
				var envs []string
				for i := range run.Env {
					envs = append(envs, run.Env[i].Name+"="+run.Env[i].Value)
				}

				fmt.Printf("Starting container %v to launch run %v %v \n", run.DockerImage, run.Name, emoji.ManTechnologist)
				container_id := RunContainer(run.DockerImage, envs)

				fmt.Printf("%v started\n\nSee logs with ( docker logs %v ) in an other terminal %v \n\n", run.DockerImage, container_id, emoji.Laptop)
				c.JSON(200, gin.H{
					"message": "ip",
				})
			})

			r.POST("/kill", func(c *gin.Context) {
				var killInstruction Kill
				if err := c.BindJSON(&killInstruction); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "bad request format",
					})
					return
				}
				fmt.Printf("Instruction to kill %v received %v \n", killInstruction.Name, emoji.ManTechnologist)
				killed := stopRunningContainer(killInstruction.Name)
				if killed {
					go getNextRun(args[1])
				}
				c.JSON(200, gin.H{
					"message": "container stop",
				})
			})

			r.POST("/terminate", func(c *gin.Context) {

				fmt.Printf("No more run to go bye bye \n")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := srv.Shutdown(ctx); err != nil {
					log.Fatal("Server forced to shutdown:", err)
				}
			})

			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
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
