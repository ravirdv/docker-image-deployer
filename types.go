package main

// deploy request json format.
type deployStruct struct {
    Name  string `json:"name,omitempty"`
    Image string `json:"image"`
    Cmd []string `json:"cmd,omitempty"`
	Env []string `json:"env_variables,omitempty"`
    Volumes []string `json:"volumes,omitempty"`
}

// error reponse type
type errorMessageStruct struct {
    ErrorMessage  string `json:"error,omitempty"`
}
