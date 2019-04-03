.. _ucloud_mysql_backup_list:

ucloud mysql backup list
------------------------

List backups of MySQL instance

Synopsis
~~~~~~~~


List backups of MySQL instance

::

  ucloud mysql backup list [flags]

Options
~~~~~~~

::

  --udb-id     string        Optional. Resource ID of UDB for list the backups of the specifid UDB 

  --backup-id     string     Optional. Resource ID of backup. List the specified backup only 

  --backup-type     string   Optional. Backup type. Accept values:auto or manual 

  --db-type     string       Optional. Only list backups of the UDB of the specified DB type 

  --begin-time     string    Optional. Begin time of backup. For example, 2019-02-26/11:21:39 

  --end-time     string      Optional. End time of backup. For example, 2019-02-26/11:31:39 

  --region     string        Optional. Override default region, see 'ucloud region' (default
                             "cn-bj2") 

  --zone     string          Optional. Override default availability zone, see 'ucloud region'
                             (default "cn-bj2-02") 

  --project-id     string    Optional. Override default project-id, see 'ucloud project list'
                             (default "org-ryrmms") 

  --offset     int           Optional. The index(a number) of resource which start to list 

  --limit     int            Optional. The maximum number of resources per page (default 100) 

  --help, -h                 help for list 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql backup <ucloud_mysql_backup>` 	 - List and manipulate backups of MySQL instance

