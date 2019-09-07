.. _ucloud_uhost_list:

ucloud uhost list
-----------------

List all UHost Instances

Synopsis
~~~~~~~~


List all UHost Instances

::

  ucloud uhost list [flags]

Options
~~~~~~~

::

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone 

  --uhost-id     strings    Optional. Resource ID of uhost instances, multiple values
                            separated by comma(without space) 

  --offset     int          Optional. Offset default 0 

  --limit     int           Optional. Limit default 50, max value 100 (default 50) 

  --group     string        Optional. Business group 

  --help, -h                help for list 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uhost <ucloud_uhost>` 	 - List,create,delete,stop,restart,poweroff or resize UHost instance

