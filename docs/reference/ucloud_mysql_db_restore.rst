.. _ucloud_mysql_db_restore:

ucloud mysql db restore
-----------------------

Create MySQL instance and restore the newly created db to the specified DB at a specified point in time

Synopsis
~~~~~~~~


Create MySQL instance and restore the newly created db to the specified DB at a specified point in time

::

  ucloud mysql db restore [flags]

Options
~~~~~~~

::

  --name     string              Required. Name of UDB instance to create 

  --src-udb-id     string        Required. Resource ID of source UDB 

  --restore-to-time     string   Required. The date and time to restore the DB to. Value must
                                 be a time in Universal Coordinated Time (UTC) format.Example:
                                 2019-02-23T23:45:00Z 

  --region     string            Optional. Override default region, see 'ucloud region'
                                 (default "cn-bj2") 

  --zone     string              Optional. Override default availability zone, see 'ucloud
                                 region' (default "cn-bj2-02") 

  --project-id     string        Optional. Override default project-id, see 'ucloud project
                                 list' (default "org-ryrmms") 

  --disk-type     string         Optional. Disk type. The default is to be consistent with the
                                 source database. Accept values: normal, ssd 

  --charge-type     string       Optional. Enumeration value.'Year',pay yearly;'Month',pay
                                 monthly; 'Dynamic', pay hourly; 'Trial', free trial(need
                                 permission) (default "Month") 

  --quantity     int             Optional. The duration of the instance. N years/months.
                                 (default 1) 

  --async, -a                    Optional. Do not wait for the long-running operation to finish 

  --help, -h                     help for restore 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql db <ucloud_mysql_db>` 	 - Manange MySQL instances

