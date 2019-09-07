.. _ucloud_mysql_conf_delete:

ucloud mysql conf delete
------------------------

Delete configuration of udb by conf-id

Synopsis
~~~~~~~~


Delete configuration of udb by conf-id

::

  ucloud mysql conf delete [flags]

Options
~~~~~~~

::

  --conf-id     string      Required. ConfID of the configuration to delete 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --help, -h                help for delete 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql conf <ucloud_mysql_conf>` 	 - List and manipulate configuration files of MySQL instances

