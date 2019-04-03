.. _ucloud_memcache_delete:

ucloud memcache delete
----------------------

Delete memcache instances

Synopsis
~~~~~~~~


Delete memcache instances

::

  ucloud memcache delete [flags]

Examples
~~~~~~~~

::

  ucloud memcache delete --umem-id umemcache-rl5xuxx/testcli1,umemcache-xsdfa/testcli2

Options
~~~~~~~

::

  --umem-id     strings     Required. Resource ID of memcache intances to delete 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region' 

  --help, -h                help for delete 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud memcache <ucloud_memcache>` 	 - List and manipulate memcache instances

