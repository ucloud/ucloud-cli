.. _ucloud_mysql_conf_list:

ucloud mysql conf list
----------------------

List configuartion files of MySQL instances

Synopsis
~~~~~~~~


List configuartion files of MySQL instances

::

  ucloud mysql conf list [flags]

Options
~~~~~~~

::

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --offset     int          Optional. The index(a number) of resource which start to list 

  --limit     int           Optional. The maximum number of resources per page (default 100) 

  --conf-id     int         Optional. Configuration identifier for the configuration to be
                            described 

  --help, -h                help for list 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql conf <ucloud_mysql_conf>` 	 - List and manipulate configuration files of MySQL instances

