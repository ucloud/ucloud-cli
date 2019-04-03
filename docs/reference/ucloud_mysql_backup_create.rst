.. _ucloud_mysql_backup_create:

ucloud mysql backup create
--------------------------

Create backups for MySQL instance manually

Synopsis
~~~~~~~~


Create backups for MySQL instance manually

::

  ucloud mysql backup create [flags]

Options
~~~~~~~

::

  --udb-id     string       Required. Resource ID of UDB instnace to backup 

  --name     string         Required. Name of backup 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --help, -h                help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql backup <ucloud_mysql_backup>` 	 - List and manipulate backups of MySQL instance

