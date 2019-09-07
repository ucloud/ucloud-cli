.. _ucloud_mysql_logs_list:

ucloud mysql logs list
----------------------

List mysql log archives(log files)

Synopsis
~~~~~~~~


List mysql log archives(log files)

::

  ucloud mysql logs list [flags]

Options
~~~~~~~

::

  --log-type     strings    Optional. Type of log. Accept Values: binlog, slow_query and error 

  --udb-id     string       Optional. Resource ID of UDB instance which the listed logs belong to 

  --begin-time     string   Optional. For example 2019-01-02/15:04:05 

  --end-time     string     Optional. For example 2019-01-02/15:04:05 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --limit     int           Optional. The maximum number of resources per page (default 100) 

  --offset     int          Optional. The index(a number) of resource which start to list 

  --help, -h                help for list 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql logs <ucloud_mysql_logs>` 	 - List and manipulate logs of MySQL instance

