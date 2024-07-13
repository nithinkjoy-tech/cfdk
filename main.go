package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/nsf/termbox-go"
	"github.com/spf13/cobra"
)

type Context struct {
	Name          string `json:"name"`
	ApplicationID string `json:"application_id"`
	Domain        string `json:"domain"`
	CompanyID     int    `json:"company_id"`
	ThemeID       string `json:"theme_id"`
	Env           string `json:"env"`
}

type Theme struct {
	ActiveContext string             `json:"active_context"`
	Contexts      map[string]Context `json:"contexts"`
}

type Config struct {
	Theme    Theme                  `json:"theme"`
	Partners map[string]interface{} `json:"partners"`
}

var selectedOption string
var config Config
var exit bool

var rootCmd = &cobra.Command{
	Use:   "cfdk",
	Short: "CLI Tool for changing FDK context",
	Run: func(cmd *cobra.Command, args []string) {
		if err := termbox.Init(); err != nil {
			log.Fatalf("failed to initialize termbox: %v", err)
		}
		defer termbox.Close()

		var err error
		config, err = readConfig(".fdk/context.json")
		if err != nil {
			log.Fatalf("failed to read config: %v", err)
		}

		options := extractUniqueDomains(config.Theme.Contexts)

		selected := 0
		for {
			printOptions(options, selected)
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyArrowUp:
					if selected > 0 {
						selected--
					}
				case termbox.KeyArrowDown:
					if selected < len(options)-1 {
						selected++
					}
				case termbox.KeyEnter:
					selectedOption = options[selected]
					updateActiveContext(selectedOption)

					if err := writeConfig(".fdk/context.json", config); err != nil {
						log.Fatalf("failed to write config: %v", err)
					}

					return
				case termbox.KeyEsc, termbox.KeyCtrlC:
					exit = true
					return
				}
			case termbox.EventError:
				log.Printf("termbox event error: %v", ev.Err)
				return
			}
		}
	},
}

func readConfig(filename string) (Config, error) {
	var config Config
	file, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	return config, err
}

func writeConfig(filename string, config Config) error {
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, configJSON, 0644)
	if err != nil {
		return err
	}
	return nil
}

func extractUniqueDomains(contexts map[string]Context) []string {
	domainSet := make(map[string]struct{})
	var domains []string
	for _, context := range contexts {
		if _, exists := domainSet[context.Domain]; !exists {
			domainSet[context.Domain] = struct{}{}
			domains = append(domains, context.Domain)
		}
	}
	return domains
}

func printOptions(options []string, selected int) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for i, option := range options {
		printOption(option, i, selected)
	}
	termbox.Flush()
}

func printOption(option string, index int, selected int) {
	x, y := 0, index
	if index == selected {
		tbprint(x, y, termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault, "> "+option)
	} else {
		tbprint(x, y, termbox.ColorDefault, termbox.ColorDefault, "  "+option)
	}
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func updateActiveContext(selectedDomain string) {
	for key, context := range config.Theme.Contexts {
		if context.Domain == selectedDomain {
			config.Theme.ActiveContext = key
			return
		}
	}
}

func runFDKLoginCommand() error {
	cmd := exec.Command("fdk", "login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runSetEnv(envName string) error {
	fmt.Println(envName)
	cmd := exec.Command("fdk", "env", "set", "-n", envName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if selectedOption != "" {
		green := "\033[32m"
		reset := "\033[0m"
		fmt.Printf("You selected: %s%s%s\n", green, selectedOption, reset)
	}

	if !exit {
		if selectedOption != "" {
			err := runSetEnv(config.Theme.Contexts[config.Theme.ActiveContext].Env)
			if err != nil {
				log.Fatalf("failed to set environment: %v", err)
			}
		}

		if err := runFDKLoginCommand(); err != nil {
			log.Fatalf("failed to run fdk login command: %v", err)
		}
	}

}
