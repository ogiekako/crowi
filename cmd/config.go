package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ogiekako/crowi/cli"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

var confCmd = &cobra.Command{
	Use:   "config",
	Short: "Config the setting file",
	Long:  "Config the setting file with your editor (default: vim)",
	RunE:  conf,
}

var (
	confGetKey  string
	confAllKeys bool
)

func conf(cmd *cobra.Command, args []string) error {
	tomlfile := cli.Conf.Core.TomlFile
	if tomlfile == "" {
		dir, _ := cli.GetDefaultDir()
		tomlfile = filepath.Join(dir, "config.toml")
	}

	config, err := toml.LoadFile(tomlfile)
	if err != nil {
		return err
	}

	if confAllKeys {
		allMap := config.ToMap()
		for _, key := range config.Keys() {
			fmt.Println(strings.Join(findKey(allMap, key), "\n"))
		}
		return nil
	}
	if confGetKey != "" {
		value := config.Get(confGetKey)
		if value != nil {
			fmt.Printf("%v\n", value)
			return nil
		}
		return fmt.Errorf("%s: no such key found", confGetKey)
	}

	editor := cli.Conf.Core.Editor
	if editor == "" {
		return errors.New("config editor not defined")
	}
	return cli.Run(editor, tomlfile)
}

func findKey(m map[string]interface{}, k string) []string {
	var ret []string
	originKey := k
	if v, ok := m[k]; ok {
		switch v.(type) {
		case map[string]interface{}:
			m = v.(map[string]interface{})
		default:
		}
	} else {
		return []string{}
	}
	for k, _ := range m {
		ret = append(ret, originKey+"."+k)
	}
	return ret
}

func init() {
	RootCmd.AddCommand(confCmd)
	confCmd.Flags().StringVarP(&confGetKey, "get", "", "", "Get the config value")
	confCmd.Flags().BoolVarP(&confAllKeys, "keys", "", false, "Get the config keys")
}
