.. _ucloud_mysql_logs_archive:

ucloud mysql logs archive
-------------------------

Archive the log of mysql as a compressed file

Synopsis
~~~~~~~~


Archive the log of mysql as a compressed file

::

  ucloud mysql logs archive [flags]

Examples
~~~~~~~~

::

  ucloud mysql logs archive --name test.cli2 --udb-id udb-xxx/test.cli1 --log-type slow_query --begin-time 2019-02-23/15:30:00 --end-time 2019-02-24/15:31:00

Options
~~~~~~~

::

  --udb-id     string       Required. Resource ID of UDB instance which we fetch logs from 

  --name     string         Required. Name of compressed file 

  --log-type     string     Required. Type of log to package. Accept values: slow_query, error 

  --begin-time     string   Optional. Required when log-type is slow. For example
                            2019-01-02/15:04:05 

  --end-time     string     Optional. Required when log-type is slow. For example
                            2019-01-02/15:04:05 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --help, -h                help for archive 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql logs <ucloud_mysql_logs>` 	 - List and manipulate logs of MySQL instance

