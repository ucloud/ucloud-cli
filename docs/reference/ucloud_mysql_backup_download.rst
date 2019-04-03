.. _ucloud_mysql_backup_download:

ucloud mysql backup download
----------------------------

Display download url of backup

Synopsis
~~~~~~~~


Display download url of backup

::

  ucloud mysql backup download [flags]

Options
~~~~~~~

::

  --backup-id     int       Required. BackupID of backup to delete (default -1) 

  --udb-id     string       Required. Resource ID of udb which the backup belongs to 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --help, -h                help for download 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql backup <ucloud_mysql_backup>` 	 - List and manipulate backups of MySQL instance

