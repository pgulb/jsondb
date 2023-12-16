package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pgulb/jsondb/config"
	"github.com/pgulb/jsondb/structures"
)

func HandleOutput(c chan structures.Response, timeOut int) ([]string, error) {
	// example function to handle response
	// set timeOut in seconds
	// jsondb responds with slice of log lines
	// if StatusOk is false, then last line will be error
	// You can write your own function if it is not suitable
	var messages []string
	select {
	case resp := <-c:
		mlen := len(resp.Message)
		for k, v := range resp.Message {
			if k != mlen-1 {
				messages = append(messages, v)
			}
		}
		if resp.StatusOk {
			messages = append(messages, resp.Message[mlen-1])
			return messages, nil
		} else {
			return messages, errors.New(resp.Message[mlen-1])
		}
	case <-time.After(time.Second * time.Duration(timeOut)):
		return nil, errors.New("no response from jsondb")
	}
}

func parseCfgArgs(cfgArgs []string, cfg map[string]string) {
	// get config path from command line arguments
	for _, v := range cfgArgs {
		if strings.Contains(v, "=") {
			splittedArg := strings.Split(v, "=")
			if splittedArg[0] == "--cfg" {
				cfg["configPath"] = splittedArg[1]
			}
		}
	}
}

func checkAndCreateJsonsPath(path string) error {
	// check path that will store jsondb files
	// create if not present
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.Mkdir(path, 0700)
		if err != nil {
			return err
		}
		return nil
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	return nil
}

func loadJsons(path string) (map[string]map[string]string, error) {
	// load jsondb files from specified location
	jsons := make(map[string]map[string]string)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() {
			fileName := filepath.Join(path, file.Name())
			bytes, err := os.ReadFile(fileName)
			if err != nil {
				return nil, err
			}
			jsonEntry := make(map[string]string)
			err = json.Unmarshal(bytes, &jsonEntry)
			if err != nil {
				return nil, err
			}
			jsons[strings.Split(file.Name(), ".")[0]] = jsonEntry
		}
	}

	// create some sample if none found
	if len(jsons) == 0 {
		entry := make(map[string]string)
		entry["_jsondbInitialKey"] = ""
		jsons["_jsondbInitialKey"] = entry
	}
	return jsons, nil
}

func saveJson(KeyFamily string, jsonValues map[string]string, path string) error {
	// dump jsondb into jsondb files (divided by KeyFamily)
	bytes, err := json.Marshal(jsonValues)
	if err != nil {
		return err
	}
	fileName := filepath.Join(path, fmt.Sprintf("%s.json", KeyFamily))
	err = os.WriteFile(fileName, bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

func handleRequest(req structures.Request,
	jsons map[string]map[string]string,
	path string) structures.Response {
	// handle request for jsondb action
	// set, get, list, listkeys
	outputMessage := []string{}
	switch req.Action {
	case "set":
		// set specific key in KeyFamily
		if req.Key == "ThisKeyWillTriggerTimeoutOnSet" {
			// key for testing purposes
			time.Sleep(time.Second * 21372137)
		}
		if len(jsons[req.KeyFamily]) == 0 {
			entry := make(map[string]string)
			entry[req.Key] = req.Value
			jsons[req.KeyFamily] = entry
		} else {
			jsons[req.KeyFamily][req.Key] = req.Value
		}
		err := saveJson(req.KeyFamily, jsons[req.KeyFamily], path)
		if err != nil {
			outputMessage = append(outputMessage, err.Error())
			return structures.Response{
				StatusOk: false,
				Message:  outputMessage,
			}
		}
		outputMessage = append(outputMessage, "\"value set\"")

	case "get":
		// get specific key in KeyFamily
		jsoned, err := json.Marshal(jsons[req.KeyFamily][req.Key])
		if err != nil {
			outputMessage = append(outputMessage, err.Error())
			return structures.Response{
				StatusOk: false,
				Message:  outputMessage,
			}
		}
		outputMessage = append(outputMessage, string(jsoned))

	case "list":
		// list KeyFamilys
		keys := []string{}
		for k := range jsons {
			keys = append(keys, k)
		}
		jsoned, err := json.Marshal(keys)
		if err != nil {
			outputMessage = append(outputMessage, err.Error())
			return structures.Response{
				StatusOk: false,
				Message:  outputMessage,
			}
		}
		outputMessage = append(outputMessage, string(jsoned))

	case "listkeys":
		// lists keys in specific keyFamily
		keys := []string{}
		for k := range jsons[req.KeyFamily] {
			keys = append(keys, k)
		}
		jsoned, err := json.Marshal(keys)
		if err != nil {
			outputMessage = append(outputMessage, err.Error())
			return structures.Response{
				StatusOk: false,
				Message:  outputMessage,
			}
		}
		outputMessage = append(outputMessage, string(jsoned))

	case "quit":
		// closes listen loop
		outputMessage = append(outputMessage, "BYE")
	}

	return structures.Response{
		StatusOk: true,
		Message:  outputMessage,
	}
}

func Listen(
	// server-like function to be invoked as goroutine
	// rest of application communicate with it using channels
	cfgArgs []string,
	input chan structures.Request,
	output chan structures.Response,
	initialResponse chan structures.Response) int {
	cfg := make(map[string]string)
	initialMessage := []string{}

	parseCfgArgs(cfgArgs, cfg)
	if cfg["configPath"] == "" {
		initialMessage = append(
			initialMessage,
			"you must provide path to config with --cfg=[[/path/to/file.json]]",
		)
		initialResponse <- structures.Response{
			StatusOk: false,
			Message:  initialMessage,
		}
		return 1
	}

	initialMessage = append(initialMessage, "reading config file...")
	jsonsPath, timeoutSeconds, err := config.ReadConfig(cfg["configPath"])
	if err != nil {
		initialMessage = append(initialMessage, err.Error())
		initialResponse <- structures.Response{
			StatusOk: false,
			Message:  initialMessage,
		}
		return 1
	}

	initialMessage = append(initialMessage, "configuration loaded succesfully")

	err = checkAndCreateJsonsPath(jsonsPath)
	if err != nil {
		initialMessage = append(initialMessage, err.Error())
		initialResponse <- structures.Response{
			StatusOk: false,
			Message:  initialMessage,
		}
		return 1
	}

	initialMessage = append(initialMessage, fmt.Sprintf("loading jsons from %s...", jsonsPath))
	jsons, err := loadJsons(jsonsPath)
	if err != nil {
		initialMessage = append(initialMessage, err.Error())
		initialResponse <- structures.Response{
			StatusOk: false,
			Message:  initialMessage,
		}
		return 1
	}
	initialMessage = append(initialMessage, "json files loaded into memory")

	initialResponse <- structures.Response{
		StatusOk: true,
		Message:  initialMessage,
	}
	close(initialResponse)

	// looping on requests
	for {
		// timeouts after number of seconds configured in config file
		// after timeout channel will dump response
		// and loop will start again
		resp := handleRequest(<-input, jsons, jsonsPath)
		if resp.Message[0] == "BYE" {
			// end listening
			select {
			case output <- resp:
			case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			}
			close(output)
			return 0
		}
		select {
		case output <- resp:
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
		}
	}
}
