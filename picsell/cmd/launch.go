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
	"github.com/spf13/viper"
)

// launchCmd represents the launch command
var launchCmd = &cobra.Command{
	Use:   "launch scan SCAN_ID",
	Short: "This launch the scan :) Then listen for next run when one is over.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		gpus := viper.GetBool("gpu")

		if len(args) < 1 {
			color.Red("Please specify the action to launch ( picsell launch scan ) \n")
			return
		}
		if args[0] == "scan" {
			if len(args) < 2 {
				color.Red("Please provide the scan ID ( picsell launch scan SCAN_ID ) \n")
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

			if resp.StatusCode == http.StatusBadRequest {
				fmt.Printf("%v  Host not registered, please run ( picsell init scan %v )  %v \n", emoji.Warning, args[1], emoji.Warning)
				return
			}

			if resp.StatusCode == http.StatusNotFound {
				fmt.Printf("%v  Scan does not exists, please check the ID  %v\n", emoji.Warning, emoji.Warning)
				return
			}

			if resp.StatusCode == http.StatusNoContent {
				fmt.Printf("%v  No more runs to launch  %v\n", emoji.Avocado, emoji.Avocado)
				return
			}

			if resp.StatusCode == http.StatusUnauthorized {
				color.Red("Wrong PICSELLIA_TOKEN")
				return
			}
			defer resp.Body.Close()

			var res Run
			json.NewDecoder(resp.Body).Decode(&res)

			var envs []string
			for i := range res.Env {
				envs = append(envs, res.Env[i].Name+"="+res.Env[i].Value)
			}

			fmt.Printf("Starting container %v to launch run %v %v \n", res.DockerImage, res.Name, emoji.ManTechnologist)
			container_id := RunContainerCmd(res.DockerImage, envs, gpus)

			if container_id == "failed" {
				log.Fatal("Container did not start")
			}
			fmt.Printf("%v started\n\nSee logs with ( docker logs %v ) in an other terminal %v \n\n", res.DockerImage, container_id, emoji.Laptop)
			// stopped := make(chan bool)

			run_id := getRunId(res)

			var port = 8080
			subdomain := getConfigHost()

			var tunnel = localtunnel.NewLocalTunnel(port)
			var errs = tunnel.OpenAs(subdomain)
			if errs != nil {
				log.Fatalf("Can not create SSH tunnel")
			}
			fmt.Printf("Training %v launched, you can visualize your performance metrics on %v https://beta.picsellia.com %v  \n\n", res.Name, emoji.Avocado, emoji.Avocado)
			color.Green("%v  Do not kill this terminal, you won't be able to perform automatical job call  %v\n", emoji.Warning, emoji.Warning)

			if run_id != "failed" {
				go checkRunning(container_id, run_id)
			}
			gin.SetMode(gin.ReleaseMode)
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
				container_id := RunContainerCmd(run.DockerImage, envs, gpus)
				fmt.Printf("%v started\n\nSee logs with ( docker logs %v ) in an other terminal %v \n\n", run.DockerImage, container_id, emoji.Laptop)
				c.JSON(200, gin.H{
					"sucess": "Next run launched",
				})
			})

			r.POST("/kill", func(c *gin.Context) {
				fmt.Printf("Received kill")
				var killInstruction Kill
				if err := c.BindJSON(&killInstruction); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "bad request format",
					})
					return
				}
				fmt.Printf("Instruction to kill %v received %v \n", killInstruction.Name, emoji.ManTechnologist)
				killed := stopRunningContainer(killInstruction.DockerImage)
				if !killed {
					fmt.Printf("Can't kill container")
				}
				c.JSON(200, gin.H{
					"message": "container stop",
				})
			})

			r.POST("/terminate", func(c *gin.Context) {

				c.JSON(200, gin.H{
					"message": "terminated",
				})
				fmt.Printf("No more run to go bye bye \n")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := srv.Shutdown(ctx); err != nil {
					log.Fatal("Server forced to shutdown:", err)
				}
			})

			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-c
				fmt.Printf("%v Exiting run, the machine will be removed from available host for your scan %v \n", emoji.Avocado, emoji.Avocado)
				killed := unregisterHost(args[1])

				time.Sleep(3000)
				if killed {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if err := srv.Shutdown(ctx); err != nil {
						log.Fatal("Server forced to shutdown:", err)
					}
					os.Exit(1)
				}

			}()
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(launchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// launchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	launchCmd.Flags().BoolP("gpu", "g", true, "Set to false if you don't have any gpu driver on your machine")
}
