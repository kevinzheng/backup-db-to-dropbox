# backup-db-to-dropbox

## Features
* Support MySQL only now
* Support one data source only now
* Clean old backup files
* Specify destination folder

## Usage
1. Put the compiled application file `backup-db-to-dropbox` in `/usr/local/bin/`.
2. `mkdir` at `/etc/backup-db-to-dropbox`.
3. `cp` `config.yaml.example` to `/etc/backup-db-to-dropbox/config.yaml`.
4. Edit `/etc/backup-db-to-dropbox/config.yaml` and set dropbox `token`, destination `folder`, and data source informations.
5.  Execute `crontab -e` and add `1 * * * * /usr/local/bin/backup-db-to-dropbox` to it, then `:wq` to save it. This will call the backing up hourly.