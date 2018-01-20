package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/ogiekako/crowi/api" // TODO (spinner)
	"github.com/crowi/go-crowi"
)

var Extention string = ".md"

type Screen struct {
	Text  string
	ID    func(string) string
	Pages *crowi.Pages
}

func NewScreen() (*Screen, error) {
	s := api.NewSpinner("Fetching...")
	s.Start()
	defer s.Stop()

	user := Conf.Crowi.User
	if user == "" {
		return &Screen{}, errors.New("user is not defined")
	}

	client, err := NewClient()
	if err != nil {
		return &Screen{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), api.Timeout)
	defer cancel()

	res, err := client.Pages.List(ctx, "", user, &crowi.PagesListOptions{
		crowi.ListOptions{Pagenation: Conf.Crowi.Paging},
	})
	if err != nil {
		return &Screen{}, err
	}

	if !res.OK {
		return &Screen{}, errors.New(res.Error)
	}

	text := ""
	for _, pi := range res.Pages {
		text += fmt.Sprintf("%s\n", pi.Path)
	}

	id := func(path string) (id string) {
		for _, pi := range res.Pages {
			if pi.Path == path {
				return pi.ID
			}
		}
		return ""
	}

	return &Screen{
		Text:  text,
		ID:    id,
		Pages: res,
	}, nil
}

type Line struct {
	Path      string
	URL       string
	ID        string
	LocalPath string
}

type Lines []Line

func (s *Screen) parseLine(line string) (*Line, error) {
	u, err := url.Parse(Conf.Crowi.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, line)
	return &Line{
		Path:      line, // for now
		URL:       u.String(),
		ID:        s.ID(line),
		LocalPath: filepath.Join(Conf.Crowi.LocalPath, s.ID(line)+Extention),
	}, nil
}

func (s *Screen) Select() (lines Lines, err error) {
	if s.Text == "" {
		err = errors.New("no text to display")
		return
	}
	selectcmd := Conf.Core.SelectCmd
	if selectcmd == "" {
		err = errors.New("no selectcmd specified")
		return
	}

	var buf bytes.Buffer
	err = runFilter(selectcmd, strings.NewReader(s.Text), &buf)
	if err != nil {
		return
	}

	if buf.Len() == 0 {
		err = errors.New("no lines selected")
		return
	}

	selectedLines := strings.Split(buf.String(), "\n")
	for _, line := range selectedLines {
		if line == "" {
			continue
		}
		var parsedLine *Line
		if parsedLine, err = s.parseLine(line); err != nil {
			return
		}
		lines = append(lines, *parsedLine)
	}

	if len(lines) == 0 {
		err = errors.New("no lines selected")
		return
	}

	return
}
