package general

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/profiles"
)

type apiServiceConfig struct {
	URL    string `json:"url"`
	ApiKey string `json:"apiKey"`
}

type apiConfig map[string]apiServiceConfig

func writeApiConfig(config apiConfig) (string, error) {
	configPath, err := profiles.GetAskConfigPath()
	if err != nil {
		return "", err
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to encode config file %s: %v", configPath, err)
	}
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return "", fmt.Errorf("failed to create config file %s: %v", configPath, err)
	}
	return configPath, nil
}

func readApiConfig() (apiConfig, error) {
	configPath, err := profiles.GetAskConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := apiConfig{
			"gemini": {
				URL:    "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent",
				ApiKey: "your_gemini_api_key_here",
			},
		}
		if _, err := writeApiConfig(defaultConfig); err != nil {
			return nil, err
		}
		fmt.Printf("Config file %s created with default values.\n", configPath)
		return defaultConfig, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check config file %s: %v", configPath, err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", configPath, err)
	}

	var config apiConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %v", configPath, err)
	}

	return config, nil
}

func fetchGeminiData(url, apiKey, query string) (string, error) {
	requestBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]any{
					{"text": query},
				},
			},
		},
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-goog-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

func NewAskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ask",
		Aliases: []string{"query", "ask-api"},
		Short:   "Ask questions to an API and get the answer",
		Long: `This command sends a query to an API (like Gemini) and gets a response. 
You can also add files to the query using the -f flag.`,
		Run: func(cmd *cobra.Command, args []string) {

			query, _ := cmd.Flags().GetString("query")
			if query == "" {
				fmt.Println("Please provide a query using -q flag.")
				return
			}

			apiConfig, err := readApiConfig()
			if err != nil {
				fmt.Println("Error reading API config:", err)
				return
			}

			service, _ := cmd.Flags().GetString("model")
			if service == "" {
				service = "gemini"
			}

			apiConfigForService, exists := apiConfig[service]
			if !exists {
				fmt.Println("Error: Service configuration for", service, "not found.")
				return
			}

			files, _ := cmd.Flags().GetStringArray("file")
			for _, file := range files {
				fileContent, err := os.ReadFile(file)
				if err != nil {
					fmt.Printf("Error reading file %s: %v\n", file, err)
					return
				}
				query += "\n" + string(fileContent)
			}

			data, err := fetchGeminiData(apiConfigForService.URL, apiConfigForService.ApiKey, query)
			if err != nil {
				fmt.Println("Error fetching data:", err)
				return
			}

			// Parsowanie JSON-a, żeby wyciągnąć tylko tekst odpowiedzi
			var resp struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
			}
			if err := json.Unmarshal([]byte(data), &resp); err != nil {
				fmt.Println("Error parsing response:", err)
				return
			}

			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				fmt.Println(resp.Candidates[0].Content.Parts[0].Text)
			} else {
				fmt.Println("No response text received.")
			}

		},
	}

	// Flags
	cmd.Flags().StringP("model", "m", "gemini", "The model to use (e.g. gemini or other_service)")
	cmd.Flags().StringP("query", "q", "", "The query to send to Gemini API")
	cmd.Flags().StringArrayP("file", "f", []string{}, "Add files to the query")

	return cmd
}
