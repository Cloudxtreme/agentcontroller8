package interceptors
import (
"errors"
"encoding/json"
"github.com/Jumpscale/agentcontroller8/core"
)

// Hashes jumpscripts executed by the jumpscript_content and store it in redis. Alters the passed command as needed

func jumpscriptInterceptor(jumpscriptStore core.JumpScriptStore) commandInterceptor {

	return func (command core.RawCommand) (core.RawCommand, error) {

		datastr, ok := command["data"].(string)
		if !ok {
			return nil, errors.New("Expecting command 'data' to be string")
		}

		data := make(map[string]interface{})
		err := json.Unmarshal([]byte(datastr), &data)
		if err != nil {
			return nil, err
		}

		content, ok := data["content"]
		if !ok {
			return nil, errors.New("jumpscript_content doesn't have content payload")
		}

		jumpscriptContent, ok := content.(string)
		if !ok {
			return nil, errors.New("Expected 'content' to be string")
		}

		id, err := jumpscriptStore.Add(core.JumpScriptContent(jumpscriptContent))

		if err != nil {
			return nil, err
		}

		//hash is stored. Now modify the command and forward it.
		delete(data, "content")
		data["hash"] = id

		updatedDatastr, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		command["data"] = string(updatedDatastr)

		return command, nil
	}
}
