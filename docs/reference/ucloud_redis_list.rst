.. _ucloud_redis_list:

ucloud redis list
-----------------

List redis instances

Synopsis
~~~~~~~~


List redis instances

::

  ucloud redis list [flags]

Options
~~~~~~~

::

  --umem-id     string      Optional. Resource ID of the redis to list 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region' 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --offset     int          Optional. The index(a number) of resource which start to list 

  --limit     int           Optional. The maximum number of resources per page (default 100) 

  --help, -h                help for list 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud redis <ucloud_redis>` 	 - List and manipulate redis instances

