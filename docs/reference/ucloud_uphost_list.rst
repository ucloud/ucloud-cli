.. _ucloud_uphost_list:

ucloud uphost list
------------------

List UPHost instances

Synopsis
~~~~~~~~


List UPHost instances

::

  ucloud uphost list [flags]

Options
~~~~~~~

::

  --help, -h                help for list 

  --limit     int           Optional. The maximum number of resources per page (default 100) 

  --offset     int          Optional. The index(a number) of resource which start to list 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --uphost-id     strings   Optional. Resource ID of uphost instances. List those specified
                            uphost instances 

  --zone     string         Optional. Override default availability zone, see 'ucloud region' 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uphost <ucloud_uphost>` 	 - List UPHost instances

