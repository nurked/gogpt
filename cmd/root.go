package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Response structure
type Response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Run function
var rootCmd = &cobra.Command{
	Use:   "gogpt",
	Short: "A CLI utility for interacting with OpenAI's GPT-3 API",
	Run: func(cmd *cobra.Command, args []string) {
		model := viper.GetString("MODEL")
		apiKey := viper.GetString("API_KEY")

		var input string

		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			reader := bufio.NewReader(os.Stdin)
			input, _ = reader.ReadString('\n')
		} else {
			if len(args) < 1 {
				fmt.Println("Please provide an input for the model.")
				os.Exit(1)
			}
			input = args[0]
		}

		message := map[string]interface{}{
			"role":    "user",
			"content": strings.TrimSuffix(input, "\n"),
		}

		reqBodyMap := map[string]interface{}{
			"model":    model,
			"messages": []map[string]interface{}{message},
		}

		reqBodyBytes, err := json.Marshal(reqBodyMap)
		if err != nil {
			fmt.Println("Error preparing the request body: ", err)
			os.Exit(1)
		}

		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBodyBytes))
		if err != nil {
			fmt.Println("Error calling the OpenAI API: ", err)
			os.Exit(1)
		}
		req.Header.Add("Authorization", "Bearer "+apiKey)
		req.Header.Add("Content-Type", "application/json")

		// Start the spinner on standard error before the request
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr)) // use the line spinner
		s.Start()                                                                                   // start the spinner

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			s.Stop() // stop the spinner
			fmt.Println("Error calling the OpenAI API: ", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		s.Stop() // stop the spinner

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading the API response: ", err)
			os.Exit(1)
		}

		var response Response
		err = json.Unmarshal(respBytes, &response)
		if err != nil {
			fmt.Println("Error parsing the API response: ", err)
			os.Exit(1)
		}

		fmt.Println(response.Choices[0].Message.Content)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(configureCmd, modelCmd)

	viper.AddConfigPath("$HOME/.gogpt")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.ReadInConfig() // Reading existing config if present
}
