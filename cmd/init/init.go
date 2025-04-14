package init

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mickamy/gotcha/internal/config"
)

const (
	path = ".gotcha.yaml"
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Generate .gotcha.yaml configuration file",
	Long:  "Generate .gotcha.yaml configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return Run()
	},
}

func Run() error {
	if _, err := os.Stat(path); err == nil {
		fmt.Print(".gotcha.yaml already exists. Overwrite? [y/N]: ")
		var res string
		if _, err := fmt.Scanln(&res); err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		if res != "y" && res != "Y" {
			fmt.Println("Canceled.")
			return nil
		}
	}

	cfg := config.Default()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println("✅ .gotcha.yaml file generated!")
	return nil
}
