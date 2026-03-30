package runner

import (
	"bytes"
	"encoding/json"
	"goclaw/agent"

	_ "embed"
)

//go:embed system.json
var defaultPrompt []byte

func getDefaultPrompt() agent.JsonObject {
	var res agent.JsonObject
	err := json.NewDecoder(bytes.NewReader(defaultPrompt)).Decode(&res)
	if err != nil {
		panic(err)
	}
	return res
}
