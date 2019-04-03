.. _ucloud_mysql_db_stop:

ucloud mysql db stop
--------------------

Stop MySQL instances by udb-id

Synopsis
~~~~~~~~


Stop MySQL instances by udb-id

::

  ucloud mysql db stop [flags]

Options
~~~~~~~

::

  --udb-id     strings      Required. Resource ID of UDB instances to stop 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --force                   Optional. Stop UDB instances by force or not 

  --async, -a               Optional. Do not wait for the long-running operation to finish. 

  --help, -h                help for stop 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql db <ucloud_mysql_db>` 	 - Manange MySQL instances

