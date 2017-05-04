package cmd

import (
	"errors"
	"path"

	"github.com/b4b4r07/crowi/cli"
	"github.com/b4b4r07/crowi/config"
	"github.com/b4b4r07/crowi/util"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open user's gist",
	Long:  "Open user's gist",
	RunE:  open,
}

func open(cmd *cobra.Command, args []string) error {
	s, err := cli.NewScreen()
	if err != nil {
		return err
	}

	selectedLines, err := util.Filter(s.Text)
	if err != nil {
		return err
	}

	if len(selectedLines) == 0 {
		return errors.New("No gist selected")
	}

	line, err := cli.ParseLine(selectedLines[0])
	if err != nil {
		return err
	}

	url := path.Join(config.Conf.Core.BaseURL, line.Path)
	return util.Open(url)
}

func init() {
	RootCmd.AddCommand(openCmd)
}
