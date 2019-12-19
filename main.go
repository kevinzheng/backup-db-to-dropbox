package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"

	uuid "github.com/nu7hatch/gouuid"
	funk "github.com/thoas/go-funk"
	"github.com/urfave/cli/v2"
)

var (
	config dropbox.Config
	folder string
)

func init() {
	var logLevel dropbox.LogLevel
	if Config.Dropbox.Log {
		logLevel = dropbox.LogInfo
	} else {
		logLevel = dropbox.LogOff
	}
	config = dropbox.Config{
		Token:    Config.Dropbox.Token,
		LogLevel: logLevel,
	}

	// "/"" is needed at start
	if !strings.HasPrefix(Config.Dropbox.Folder, "/") {
		folder = "/" + Config.Dropbox.Folder
	} else {
		folder = Config.Dropbox.Folder
	}
}

func dump() (string, error) {
	app := "mysqldump"

	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(Config.Backup.TmpDir, u.String())
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
	}

	t := time.Now()
	timestamp := t.Format(Config.Backup.FilenameTimeForamt)
	var filename = fmt.Sprintf("%s-%s.sql", strings.Join(Config.Source.Dbs, "-"), timestamp)

	p := filepath.Join(dir, filename)

	var args []string
	args = append(args, "-u"+Config.Source.Username)

	if len(Config.Source.Password) > 0 {
		args = append(args, "-p"+Config.Source.Password)
	}

	args = append(args, "--databases")
	args = append(args, Config.Source.Dbs...)

	cmd := exec.Command(app, args...)

	outfile, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error: %v", stderr.String())
	}

	return p, nil
}

func compress(path string) (string, error) {
	tarFile := path + ".tar.gz"

	app := "tar"
	args := []string{"-C", filepath.Dir(path), "-czvf", tarFile, filepath.Base(path)}
	cmd := exec.Command(app, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error: %v", stderr.String())
	}
	return tarFile, nil
}

func upload(tarFile string) error {
	dbx := files.New(config)

	file, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer file.Close()

	listFolderResult, err := dbx.ListFolder(&files.ListFolderArg{
		Path: "", // Specify the root folder as an empty string rather than as "/".
	})
	if err != nil {
		return err
	}

	folderExisted := funk.Find(listFolderResult.Entries, func(v files.IsMetadata) bool {
		metadata := v.(*files.FolderMetadata)
		return metadata.Name == filepath.Base(folder)
	})

	if folderExisted == nil {
		_, err := dbx.CreateFolderV2(&files.CreateFolderArg{
			Path: folder,
		})
		if err != nil {
			return err
		}
	}

	f, err := os.Open(tarFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	targetPath := filepath.Join(folder, filepath.Base(tarFile))

	var r io.Reader = f
	_, err = dbx.Upload(&files.CommitInfo{
		Path: targetPath,
		Mode: &files.WriteMode{
			Tagged: dropbox.Tagged{
				Tag: "add",
			},
		},
	}, r)
	if err != nil {
		return err
	}

	fmt.Printf(">>>>>>> uploaded %s\n", filepath.Base(tarFile))

	return nil
}

func clean(folder string) error {
	return os.RemoveAll(folder)
}

func removeOldBackup() error {
	dbx := files.New(config)

	listFolderResult, err := dbx.ListFolder(&files.ListFolderArg{
		Path: folder,
	})
	if err != nil {
		return err
	}

	for _, entity := range listFolderResult.Entries {
		metadata := entity.(*files.FileMetadata)
		d := time.Duration(time.Duration(Config.Backup.KeepDays) * 24 * time.Hour)
		if time.Time.Before(metadata.ClientModified, time.Now().Truncate(d)) {
			_, err := dbx.DeleteV2(&files.DeleteArg{
				Path: metadata.PathLower,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "backup-db-to-dropbox",
		Usage: "./backup-db-to-dropbox",
		Action: func(c *cli.Context) error {
			p, err := dump()
			if err != nil {
				panic(err)
			}
			tarFile, err := compress(p)
			if err != nil {
				panic(err)
			}
			err = upload(tarFile)
			if err != nil {
				panic(err)
			}
			err = clean(filepath.Dir(p))
			if err != nil {
				panic(err)
			}
			err = removeOldBackup()
			if err != nil {
				panic(err)
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
