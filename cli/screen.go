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
	crowi "github.com/ogiekako/go-growi"
)

var Extention string = ".md"

type Screen struct {
	Text  string
	ID    func(string) (id, path string)
	Pages *crowi.Pages
}

func NewScreen(path string) (*Screen, error) {
	s := api.NewSpinner("Fetching...")
	s.Start()
	defer s.Stop()

	user := Conf.Crowi.User
	if user == "" {
		return &Screen{}, errors.New("user is not defined")
	}
	if path != "" {
		user = ""
	}

	client, err := NewClient()
	if err != nil {
		return &Screen{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), api.Timeout)
	defer cancel()

	res, err := client.Pages.List(ctx, path, user, &crowi.PagesListOptions{
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
		s, err := url.PathUnescape(pi.Path)
		if err != nil {
			return nil, err
		}
		text += fmt.Sprintf("%s\n", s)
	}

	id := func(unescapedPath string) (id, path string) {
		for _, pi := range res.Pages {
			if pi.Path == unescapedPath {
				return pi.ID, pi.Path
			}
			p, err := url.PathUnescape(pi.Path)
			if err != nil {
				continue
			}
			if p == unescapedPath {
				return pi.ID, pi.Path
			}
		}
		return "", ""
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
	id, p := s.ID(line)
	if id == "" {
		return nil, fmt.Errorf("Failed to find ID for %q", line)
	}
	u, err := url.Parse(Conf.Crowi.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, line)
	res := &Line{
		Path:      p, // for now
		URL:       u.String(),
		ID:        id,
		LocalPath: filepath.Join(Conf.Crowi.LocalPath, id+Extention),
	}
	return res, nil
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
	return s.lines(buf)
}

func (s *Screen) GetAll() (lines Lines, err error) {
	var buf bytes.Buffer
	if n, _ := buf.WriteString(s.Text); n == 0 {
		err = errors.New("no lines selected")
		return
	}
	return s.lines(buf)
}

func (s *Screen) lines(buf bytes.Buffer) (lines Lines, err error) {
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
