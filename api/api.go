package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	crowi "github.com/ogiekako/go-growi"
)

var (
	Timeout time.Duration = 30 * time.Second
)

type Page struct {
	Info      crowi.PageInfo
	LocalPath string
	Client    *crowi.Client
}

func NewPage(client *crowi.Client) *Page {
	if client == nil {
		client = &crowi.Client{}
	}
	return &Page{
		Info:      crowi.PageInfo{},
		LocalPath: "",
		Client:    client,
	}
}

func (page Page) Create(path, body string) (*crowi.Page, error) {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	s := NewSpinner("Posting...")
	s.Start()
	defer s.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	return page.Client.Pages.Create(ctx, path, body)
}

func fileContent(file string) (string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (page Page) upload() (res *crowi.Page, err error) {
	s := NewSpinner("Uploading...")
	s.Start()
	defer s.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	res, err = page.getWithUnescaping(ctx, page.Info.Path)
	if err != nil {
		return nil, fmt.Errorf("upload: failed to get: %w", err)
	}
	if res.Error != "" {
		return nil, fmt.Errorf("upload: Get failed: %v", res.Error)
	}

	remoteBody := res.Page.Revision.Body
	localBody, err := fileContent(page.LocalPath)
	if err != nil {
		return
	}

	if remoteBody == localBody {
		// do nothing
		return &crowi.Page{}, nil
	}

	// Empty body causes error: page_id and body are required.
	if localBody == "" {
		localBody = "."
	}

	res, err = page.Client.Pages.Update(ctx, page.Info.ID, res.Page.RevisionID, localBody)
	if err != nil {
		return res, fmt.Errorf("upload: failed to update: %w", err)
	}
	return
}

func (page Page) getWithUnescaping(ctx context.Context, p string) (*crowi.Page, error) {
	if q, err := url.PathUnescape(p); err == nil {
		p = q
	}
	return page.Client.Pages.Get(ctx, p)
}

func (page Page) download() (res *crowi.Page, err error) {
	s := NewSpinner("Downloading...")
	s.Start()
	defer s.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	res, err = page.getWithUnescaping(ctx, page.Info.Path)
	if err != nil {
		return
	}
	if res.Error != "" {
		return res, fmt.Errorf("download: failed to Get: %s", res.Error)
	}

	remoteBody := res.Page.Revision.Body
	localBody, err := fileContent(page.LocalPath)
	switch {
	case os.IsNotExist(err):
		// Fall through.
	case err != nil:
		return nil, err
	case remoteBody == localBody:
		// do nothing
		return &crowi.Page{}, nil
	}

	err = ioutil.WriteFile(page.LocalPath, []byte(remoteBody), os.ModePerm)
	return
}

func exists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func (page Page) Sync() (err error) {
	var res *crowi.Page

	if !exists(page.LocalPath) {
		res, err = page.download()
		if res.OK {
			fmt.Printf("Downloaded %s\n", res.Page.Path)
		}
		return err
	}
	fi, err := os.Stat(page.LocalPath)
	if err != nil {
		return err
	}

	local := fi.ModTime().UTC()
	remote := page.Info.UpdatedAt.UTC()

	switch {
	case local.After(remote):
		res, err = page.upload()
		if err != nil {
			return fmt.Errorf("Failed upload: %w", err)
		}
		if res.OK {
			fmt.Printf("Uploaded %s\n", res.Page.Path)
		} else if res.Error != "" {
			return fmt.Errorf("Failed to upload: %s", res.Error)
		}
	case remote.After(local):
		res, err = page.download()
		if res.OK {
			fmt.Printf("Downloaded %s\n", res.Page.Path)
		}
	default:
	}

	return err
}

func (page Page) Attach(id, file string) (*crowi.Attachment, error) {
	return page.Client.Attachments.Add(context.Background(), id, file)
}

func (page Page) Images(id string) (*crowi.Attachments, error) {
	return page.Client.Attachments.List(context.Background(), id)
}

func (page Page) Update(id, revisionID, body string) (res *crowi.Page, err error) {
	return page.Client.Pages.Update(context.Background(), id, revisionID, body)
}
