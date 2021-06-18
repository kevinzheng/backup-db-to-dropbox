package app

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/fatih/color"
	"springup.xyz/backupdbtodropbox/config"

	"github.com/go-co-op/gocron"

	uuid "github.com/nu7hatch/gouuid"
	funk "github.com/thoas/go-funk"
)

func dumpMySQL() (string, error) {
	app := "mysqldump"

	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(config.Config.Backup.TmpDir, u.String())
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
	}

	t := time.Now()
	timestamp := t.Format(config.Config.Backup.FilenameTimeForamt)
	var filename = fmt.Sprintf("%s%s-%s.sql", config.Config.Backup.Prefix, strings.Join(config.Config.Source.Dbs, "-"), timestamp)

	p := filepath.Join(dir, filename)

	var args []string
	args = append(args, "--host="+config.Config.Source.Host)
	args = append(args, "--port="+config.Config.Source.Port)
	args = append(args, "--user="+config.Config.Source.Username)

	if len(config.Config.Source.Password) > 0 {
		args = append(args, "--password="+config.Config.Source.Password)
	}

	args = append(args, "--databases")
	args = append(args, config.Config.Source.Dbs...)

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

func dumpPostgres() (string, error) {
	app := "pg_dump"

	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(config.Config.Backup.TmpDir, u.String())
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
	}

	t := time.Now()
	timestamp := t.Format(config.Config.Backup.FilenameTimeForamt)
	var filename = fmt.Sprintf("%s%s-%s.sql", config.Config.Backup.Prefix, strings.Join(config.Config.Source.Dbs, "-"), timestamp)

	p := filepath.Join(dir, filename)

	var args []string
	args = append(args, "--host="+config.Config.Source.Host)
	args = append(args, "--port="+config.Config.Source.Port)
	args = append(args, "--username="+config.Config.Source.Username)

	args = append(args, "-d")
	args = append(args, config.Config.Source.Dbs...)

	cmd := exec.Command(app, args...)

	if len(config.Config.Source.Password) > 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "PGPASSWORD="+config.Config.Source.Password)
	}

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

	fmt.Printf(">>>>>>> compressed backup file %s\n", tarFile)

	return tarFile, nil
}

func upload(dropboxConfig dropbox.Config, folder string, tarFile string) error {
	dbx := files.New(dropboxConfig)

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

	file, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer file.Close()

	targetPath := filepath.Join(folder, filepath.Base(tarFile))

	var r io.Reader = file
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

func removeOldBackup(dropboxConfig dropbox.Config, folder string) error {
	dbx := files.New(dropboxConfig)

	listFolderResult, err := dbx.ListFolder(&files.ListFolderArg{
		Path: folder,
	})
	if err != nil {
		return err
	}

	for _, entity := range listFolderResult.Entries {
		metadata := entity.(*files.FileMetadata)
		d := time.Duration(time.Duration(config.Config.Backup.KeepDays) * 24 * time.Hour)
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

func Backup(dropboxConfig dropbox.Config, folder string) {
	var p string
	var err error
	if config.Config.Source.Type == "mysql" {
		p, err = dumpMySQL()
		if err != nil {
			panic(err)
		}
	} else if config.Config.Source.Type == "postgres" {
		p, err = dumpPostgres()
		if err != nil {
			panic(err)
		}
	}
	tarFile, err := compress(p)
	if err != nil {
		panic(err)
	}
	err = upload(dropboxConfig, folder, tarFile)
	if err != nil {
		panic(err)
	}
	err = clean(filepath.Dir(p))
	if err != nil {
		panic(err)
	}
	err = removeOldBackup(dropboxConfig, folder)
	if err != nil {
		panic(err)
	}

	color.Green(fmt.Sprintf("Backed up database %s to Dropbox at %s...", strings.Join(config.Config.Source.Dbs, ", "), time.Now().String()))
}

func Schedule(cronExpression string, dropboxConfig dropbox.Config, folder string) {
	s := gocron.NewScheduler(time.UTC)
	s.Cron(cronExpression).Do(Backup, dropboxConfig, folder)
	s.StartBlocking()
}
