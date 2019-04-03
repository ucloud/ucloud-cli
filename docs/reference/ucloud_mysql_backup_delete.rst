.. _ucloud_mysql_backup_delete:

ucloud mysql backup delete
--------------------------

Delete backups of MySQL instance

Synopsis
~~~~~~~~


Delete backups of MySQL instance

::

  ucloud mysql backup delete [flags]

Examples
~~~~~~~~

::

  ucloud udb backup delete --backup-id 65534,65535

Options
~~~~~~~

::

  --backup-id     ints      Required. BackupID of backups to delete 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --help, -h                help for delete 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql backup <ucloud_mysql_backup>` 	 - List and manipulate backups of MySQL instance

