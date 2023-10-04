package snake

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type PlayersControlSettings struct {
	PlayersControlSettings []PlayerControlSetting `json:"playerControlSetting"`
}

type PlayerControlSetting struct {
	Color string `json:"color"`
	Up    string `json:"up"`
	Down  string `json:"down"`
	Right string `json:"right"`
	Left  string `json:"left"`
}

func controlls(fileName string) PlayersControlSettings {
	// Open our jsonFile
	jsonFile, err := os.Open(fileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened control.json")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var playerControlSettings PlayersControlSettings
	json.Unmarshal(byteValue, &playerControlSettings)

	// for i := 0; i < len(playerControlSettings.PlayersControlSettings); i++ {
	// 	fmt.Println(playerControlSettings.PlayersControlSettings[i].Color)
	// 	fmt.Println(playerControlSettings.PlayersControlSettings[i].Up)
	// 	fmt.Println(playerControlSettings.PlayersControlSettings[i].Down)
	// 	fmt.Println(playerControlSettings.PlayersControlSettings[i].Right)
	// 	fmt.Println(playerControlSettings.PlayersControlSettings[i].Left)
	// 	fmt.Println("--")
	// }
	return playerControlSettings
}
