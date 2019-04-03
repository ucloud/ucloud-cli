.. _ucloud_mysql_db_start:

ucloud mysql db start
---------------------

Start MySQL instances by udb-id

Synopsis
~~~~~~~~


Start MySQL instances by udb-id

::

  ucloud mysql db start [flags]

Options
~~~~~~~

::

  --udb-id     strings      Required. Resource ID of UDB instances to start 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --async, -a               Optional. Do not wait for the long-running operation to finish. 

  --help, -h                help for start 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql db <ucloud_mysql_db>` 	 - Manange MySQL instances

