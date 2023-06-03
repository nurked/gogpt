package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Model struct {
	ID string `json:"id"`
}

type ModelResponse struct {
	Data []Model `json:"data"`
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Interactively configure the API Key and model",
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter API Key: ")
		apiKey, _ := reader.ReadString('\n')

		models, err := fetchModels(apiKey)
		if err != nil {
			fmt.Printf("Error fetching models: %s", err)
			os.Exit(1)
		}

		fmt.Println("Available models:")
		for i, model := range models {
			fmt.Printf("%d: %s\n", i+1, model.ID)
		}

		fmt.Print("Select model by number: ")
		input, _ := reader.ReadString('\n')
		selected, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || selected < 1 || selected > len(models) {
			fmt.Println("Invalid selection")
			os.Exit(1)
		}

		viper.Set("API_KEY", strings.TrimSpace(apiKey))
		viper.Set("MODEL", models[selected-1].ID)

		err = viper.WriteConfig()
		if err != nil {
			fmt.Printf("Error saving config: %s", err)
			os.Exit(1)
		}
	},
}

func fetchModels(apiKey string) ([]Model, error) {
	req, err := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+strings.TrimSpace(apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var modelResp ModelResponse
	err = json.NewDecoder(resp.Body).Decode(&modelResp)
	if err != nil {
		return nil, err
	}

	return modelResp.Data, nil
}

func init() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	configPath := filepath.Join(dir, ".gogpt")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.Mkdir(configPath, 0755)
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			viper.SafeWriteConfigAs(filepath.Join(configPath, "config.yaml"))
		} else {
			// Config file was found but another error was produced
			fmt.Printf("Fatal error config file: %s", err)
			os.Exit(1)
		}
	}
}
