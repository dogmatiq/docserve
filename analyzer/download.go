package analyzer

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

func downloadAndUncompress(
	ctx context.Context,
	url, dir string,
	c *http.Client,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	if c == nil {
		c = http.DefaultClient
	}

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	uncompressed, err := gzip.NewReader(res.Body)
	if err != nil {
		return err
	}
	defer uncompressed.Close()

	archive := tar.NewReader(uncompressed)
	trimPrefix := ""

	for {
		header, err := archive.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(header.Name, trimPrefix)
		entryPath := path.Join(dir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if trimPrefix == "" {
				// GitHub puts the repository content inside another top-level
				// directory within the archive. We don't want this directory,
				// so we move the content of the archive "up one level" by
				// trimming this directory name from the start of every file.
				trimPrefix = header.Name
				continue
			}

			if err := os.Mkdir(entryPath, 0700); err != nil {
				return err
			}

		case tar.TypeReg:
			f, err := os.OpenFile(entryPath, os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, archive); err != nil {
				return err
			}
		}
	}
}
