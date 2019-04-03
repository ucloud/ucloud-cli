.. _ucloud_mysql_conf_update:

ucloud mysql conf update
------------------------

Update parameters of DB's configuration

Synopsis
~~~~~~~~


Update parameters of DB's configuration

::

  ucloud mysql conf update [flags]

Options
~~~~~~~

::

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --conf-id     string      Required. ConfID of configuration to update 

  --key     string          Optional. Key of parameter 

  --value     string        Optional. Value of parameter 

  --file     string         Optional. Path of file in which each parameter occupies one line
                            with format 'key = value' 

  --help, -h                help for update 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql conf <ucloud_mysql_conf>` 	 - List and manipulate configuration files of MySQL instances

