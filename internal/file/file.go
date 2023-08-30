package file

import (
	"archive/zip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func CopyFile(src, dst string) (err error) {
	// Reference - https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	// Create directory for target file if it doesn't exist yet
	err = os.MkdirAll(filepath.Dir(dst), os.ModePerm)
	if err != nil {
		return
	}

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && errors.Is(err, fs.ErrExist) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Type() == fs.ModeSymlink {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
func UnzipSource(source, destination string) (err error) {
	// 1. Open the zip file
	reader, err := zip.OpenReader(source)
	if err != nil {
		return
	}
	defer reader.Close()

	// 2. Get the absolute destination path
	destination, err = filepath.Abs(destination)
	if err != nil {
		return
	}

	// 3. Iterate over zip files inside the archive and unzip each of them
	for _, f := range reader.File {
		err = unzipFile(f, destination)
		if err != nil {
			return
		}
	}
	return
}

func unzipFile(f *zip.File, destination string) (err error) {
	// 4. Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// 5. Create directory tree
	if f.FileInfo().IsDir() {
		if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
			return
		}
		return
	}

	if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return
	}

	// 6. Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return
	}
	defer destinationFile.Close()

	// 7. Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return
	}
	defer zippedFile.Close()

	if _, err = io.Copy(destinationFile, zippedFile); err != nil {
		return
	}
	return
}

func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

func ReplaceDir(src string, dst string) (err error) {
	err = os.RemoveAll(dst)
	if err != nil {
		return
	}
	return copyDir(src, dst)
}

// ZipDir Compress a directory into a zip file
func ZipDir(src string, dst string, includeSrc bool) error {
	zipfile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var name string
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if includeSrc {
			name, err = filepath.Rel(filepath.Dir(src), path)
			if err != nil {
				return err
			}
		} else {
			if path == src {
				return nil
			}
			name, err = filepath.Rel(src, path)
			if err != nil {
				return err
			}
		}

		header.Name = filepath.ToSlash(name)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})

	return err
}

func ZipDirToBase64(src string) (string, error) {
	zipFile := src + ".zip"
	err := ZipDir(src, zipFile, false)
	if err != nil {
		return "", err
	}
	fileContent, err := os.ReadFile(zipFile)
	if err != nil {
		return "", err
	}
	err = os.Remove(zipFile)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(fileContent), nil
}
