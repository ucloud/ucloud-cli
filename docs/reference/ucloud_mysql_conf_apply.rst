.. _ucloud_mysql_conf_apply:

ucloud mysql conf apply
-----------------------

Apply configuration for UDB instances

Synopsis
~~~~~~~~


Apply configuration for UDB instances

::

  ucloud mysql conf apply [flags]

Options
~~~~~~~

::

  --conf-id     string        Required. ConfID of the configuration to be applied 

  --udb-id     strings        Required. Resource ID of UDB instances to change configuration 

  --restart-after-apply       Optional. The new configuration will take effect after DB
                              restarts (default true) 

  --yes, -y                   Optional. Do not prompt for confirmation 

  --async, -a                 Optional. Do not wait for the long-running operation to finish. 

  --region     string         Optional. Override default region, see 'ucloud region' (default
                              "cn-bj2") 

  --zone     string           Optional. Override default availability zone, see 'ucloud
                              region' (default "cn-bj2-02") 

  --project-id     string     Optional. Override default project-id, see 'ucloud project list'
                              (default "org-ryrmms") 

  --help, -h                  help for apply 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql conf <ucloud_mysql_conf>` 	 - List and manipulate configuration files of MySQL instances

