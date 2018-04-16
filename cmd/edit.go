package cmd

import (
	"github.com/ogiekako/crowi/cli"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [path]",
	Short: "Edit a page",
	Long:  `Edit a page`,
	RunE:  edit,
}

func edit(cmd *cobra.Command, args []string) error {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	screen, err := cli.NewScreen(path)
	if err != nil {
		return err
	}
	var lines cli.Lines
	if path == "" {
		lines, err = screen.Select()
		if err != nil {
			return err
		}
	} else {
		lines, err = screen.GetAll()
		if err != nil {
			return err
		}
	}
	return cli.EditPage(screen.Pages, lines[0])
}

func init() {
	RootCmd.AddCommand(editCmd)
}
