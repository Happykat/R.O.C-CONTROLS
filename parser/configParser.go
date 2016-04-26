package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Cmd struct {
	Name              string
	Code              byte
	Min, Max, Default int
}

type Cmds struct {
	Commands []Cmd
}

func RobotCommand(f_path string) (map[string]Cmd, error) {

	var cmds Cmds

	c := make(map[string]Cmd)
	data, err := ioutil.ReadFile(f_path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	err = json.Unmarshal(data, &cmds)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for _, kk := range cmds.Commands {
		c[kk.Name] = kk
	}
	fmt.Println(c)
	return c, nil
}