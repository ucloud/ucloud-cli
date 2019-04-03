.. _ucloud_mysql_db_list:

ucloud mysql db list
--------------------

List MySQL instances

Synopsis
~~~~~~~~


List MySQL instances

::

  ucloud mysql db list [flags]

Options
~~~~~~~

::

  --udb-id     string       Optional. List the specified mysql 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --limit     int           Optional. The maximum number of resources per page (default 100) 

  --offset     int          Optional. The index(a number) of resource which start to list 

  --include-slaves          Optional. When specifying the udb-id, whether to display its
                            slaves together. Accept values:true, false 

  --help, -h                help for list 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql db <ucloud_mysql_db>` 	 - Manange MySQL instances

