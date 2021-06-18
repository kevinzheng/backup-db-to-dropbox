package config

var (
	Config struct {
		Backup struct {
			TmpDir             string `yaml:"tmpDir"`
			FilenameTimeForamt string `yaml:"filenameTimeForamt"`
			KeepDays           int    `yaml:"keepDays"`
			Prefix             string `yaml:"prefix"`
			Cron               string `yaml:"cron"`
		}

		Dropbox struct {
			Token  string `yaml:"token"`
			Log    bool   `yaml:"log"`
			Folder string `yaml:"folder"`
		}

		Source struct {
			Type     string   `yaml:"type"`
			Host     string   `yaml:"host"`
			Port     string   `yaml:"port"`
			Username string   `yaml:"username"`
			Password string   `yaml:"password"`
			Dbs      []string `yaml:"dbs"`
		} `yaml:"source"`
	}
)
