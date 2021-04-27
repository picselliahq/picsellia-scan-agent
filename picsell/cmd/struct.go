package cmd

type EntireRun struct {
	CreationDate string                 `json:"date_created"`
	ID           string                 `json:"id"`
	Parameters   map[string]interface{} `json:"parameters"`
	Experiment   ExperimentStruct       `json:"experiment"`
	Metric       string                 `json:"metric"`
	Value        float32                `json:"value"`
	Duration     int                    `json:"duration"`
	StartTime    string                 `json:"start_time"`
	EndTime      string                 `json:"end_time"`
	Order        int                    `json:"order"`
	Status       string                 `json:"status"`
	Info         map[string]interface{} `json:"info"`
	DockerImage  string                 `json:"docker_image"`
	Env          []EnvStruc             `json:"env"`
}

type Run struct {
	Name        string     `json:"name"`
	DockerImage string     `json:"docker_image"`
	Env         []EnvStruc `json:"env"`
}

type Kill struct {
	Name        string `json:"name"`
	DockerImage string `json:"docker_image"`
}

type ExperimentStruct struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type EnvStruc struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

type DockerRun struct {
	DockerImage string `json:"docker_image"`
}
