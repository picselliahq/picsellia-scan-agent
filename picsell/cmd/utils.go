package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/enescakir/emoji"
	uuid "github.com/satori/go.uuid"

	//"io"
	"io/ioutil"
	"os"
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
		fmt.Println(string(file))
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
		panic(err)
	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return resp.ID
}

func getAPIToken() string {

	api_token, err := os.LookupEnv("PICSELLIA_TOKEN")

	if !err {
		panic("Please set up your PICSELLIA_TOKEN in your env variable ( export PICSELLIA_TOKEN=your_token ) ")
	}

	return api_token
}

func unregisterHost(sweepId string) {

	host_id := getConfigHost()

	token := getAPIToken()
	var bearer = "Token " + token

	req, err := http.NewRequest("POST", URL+"sweep/"+sweepId+"/"+host_id+"/unregister", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	_, err = client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
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
			fmt.Print("Stopping container ", container.ID[:10], "...")
			if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
				panic(err)
			}
			fmt.Printf("%v Stopped", imageName)
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
