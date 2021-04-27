package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
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
