package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"strings"

	"github.com/enescakir/emoji"
	uuid "github.com/satori/go.uuid"

	//"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func getConfigHost() string {
	user, err := user.Current()

	if err != nil {
		panic(err)
	}
	path := filepath.Join(user.HomeDir, ".picsell_config")
	_, err = os.Stat(path)
	if err == nil {
		file, _ := ioutil.ReadFile(path)
		conf := Configuration{}
		_ = json.Unmarshal([]byte(file), &conf)
		return conf.TunnelUrlID

	} else {
		fmt.Printf("First use of Picsell CLI on this machine, creating Host record for SSH communication ..\n")
		uuid := uuid.NewV4()
		uuidStr := uuid.String()
		_, err = os.Create(path)
		if err != nil {
			fmt.Printf("Can't create config file at %v", path)
			panic(err)
		}
		conf := Configuration{
			TunnelUrlID: uuidStr,
		}
		file, _ := json.Marshal(&conf)
		_ = ioutil.WriteFile(path, file, 0644)
		return uuidStr
	}

}

func RunContainer(imageName string, envs []string) string {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{Image: imageName, Env: envs}, nil, nil, nil, "")
	if err != nil {
		log.Fatalf("Container %v does not exists, please run picsell init scan ID first\nIf the problem persist, try running (docker pull %v) and retry",
			imageName, imageName)

	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return resp.ID
}

func RunContainerCmd(imageName string, envs []string, gpus bool) string {
	app := "docker"

	arg0 := "run"
	arg4 := imageName

	arg1 := "--gpus all"

	var tmp string

	for i := range envs {
		tmp += envs[i] + "\n"
	}

	bString := []byte(tmp)

	f, err := os.Create(".tmpenv")
	if err != nil {
		log.Fatal("Can't write file")
	}

	defer f.Close()

	_, err = f.Write(bString)
	if err != nil {
		log.Fatal("Can't write file")
	}

	arg2 := "--env-file=.tmpenv"
	arg3 := "-d"

	if gpus {
		cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4)
		_, err = cmd.Output()
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		cmd := exec.Command(app, arg0, arg2, arg3, arg4)
		_, err = cmd.Output()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {

		if container.Image == imageName {
			return container.ID
		}
	}

	return "failed"
}

func getAPIToken() string {

	api_token, err := os.LookupEnv("PICSELLIA_TOKEN")

	if !err {
		log.Fatalf("Please set up your PICSELLIA_TOKEN in your env variable ( export PICSELLIA_TOKEN=your_token ) \n")
	}

	return api_token
}

func unregisterHost(sweepId string) bool {

	host_id := getConfigHost()

	token := getAPIToken()
	var bearer = "Token " + token

	req, err := http.NewRequest("POST", URL+"sweep/"+sweepId+"/"+host_id+"/unregister", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Successfully unregistered host \n")
	} else {
		fmt.Println(resp.StatusCode)
	}

	return true
}

func stopRunningContainer(imageName string) bool {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {

		if container.Image == imageName {
			fmt.Print("Stopping container ", container.ID[:10], "... \n")
			if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
				panic(err)
			}
			fmt.Printf("%v Stopped \n", imageName)
			return true
		}
	}
	return false
}

func getNextRun(sweepID string) {

	host_id := getConfigHost()
	url := URL + "run/cli/" + sweepID + "/next_run/" + host_id
	token := getAPIToken()
	var bearer = "Token " + token

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%v  Picsell platform not accessible, please try again later  %v \n", emoji.Warning, emoji.Warning)
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
	container_id := RunContainer(res.DockerImage, envs)
	fmt.Printf("%v started\n\nSee logs with ( docker logs %v ) in an other terminal %v \n\n", res.DockerImage, container_id, emoji.Laptop)

}

func indicator(shutdownCh <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Print(".")
		case <-shutdownCh:
			return
		}
	}
}

func checkRunning(container_id string, run_id string) {

	for {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Fatalf("Docker Daemon stopped")
		}

		options := types.ContainerLogsOptions{ShowStdout: true}
		_, err = cli.ContainerLogs(ctx, container_id, options)

		if err != nil {
			if !strings.Contains(err.Error(), "too many open files") {
				token := getAPIToken()
				var bearer = "Token " + token

				req, err := http.NewRequest("POST", URL+"run/cli/possible_failure/"+run_id, nil)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Add("Authorization", bearer)

				client := &http.Client{}
				resp, err := client.Do(req)

				if err != nil {
					log.Fatalf("%v  Picsell platform not accessible, please try again later  %v \n", emoji.Warning, emoji.Warning)
					return
				}

				if resp.StatusCode == http.StatusNotFound {
					log.Fatalf("%v  Run does not exists  %v \n", emoji.Warning, emoji.Warning)
					return
				}
				if resp.StatusCode == http.StatusOK {
					return
				}
				if resp.StatusCode == http.StatusCreated {
					log.Fatalf("%v  Docker image failed, inspect the logs with (docker logs %v )  %v \n", emoji.Warning, container_id, emoji.Warning)
					return
				}
			} else {
			}
			//panic(err)

		}
	}
}

func getRunId(run Run) string {

	for i := range run.Env {
		if run.Env[i].Name == "run_id" {
			return run.Env[i].Value
		}
	}

	return "failed"
}
