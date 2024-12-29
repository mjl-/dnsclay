package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

func cmdLicense(args []string) {
	if len(args) != 0 {
		flag.Usage()
	}

	err := licenseWrite(os.Stdout)
	xcheckf(err, "write")
}

func licenseWrite(dst io.Writer) error {
	copyFile := func(p string) error {
		f, err := fsys.Open(p)
		if err != nil {
			return fmt.Errorf("open license file: %v", err)
		}
		if _, err := io.Copy(dst, f); err != nil {
			return fmt.Errorf("copy license file: %v", err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("close license file: %v", err)
		}
		return nil
	}

	err := fs.WalkDir(fsys, "license", func(path string, d fs.DirEntry, err error) error {
		if !d.Type().IsRegular() {
			return nil
		}
		if _, err := fmt.Fprintf(dst, "\n\n# %s\n\n", strings.TrimPrefix(path, "license/")); err != nil {
			return err
		}
		return copyFile(path)
	})
	if err != nil {
		return fmt.Errorf("walk license: %v", err)
	}
	return nil
}
