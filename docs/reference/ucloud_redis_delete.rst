.. _ucloud_redis_delete:

ucloud redis delete
-------------------

Delete redis instances

Synopsis
~~~~~~~~


Delete redis instances

::

  ucloud redis delete [flags]

Examples
~~~~~~~~

::

  ucloud redis delete --umem-id uredis-rl5xuxx/testcli1,uredis-xsdfa/testcli2

Options
~~~~~~~

::

  --umem-id     strings     Required. Resource ID of redis intances to delete 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --help, -h                help for delete 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud redis <ucloud_redis>` 	 - List and manipulate redis instances

