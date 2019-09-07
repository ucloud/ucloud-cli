.. _ucloud_ulb_list:

ucloud ulb list
---------------

List ULB instances

Synopsis
~~~~~~~~


List ULB instances

::

  ucloud ulb list [flags]

Options
~~~~~~~

::

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --ulb-id     string       Optional. Resource ID of ULB instance to list 

  --vpc-id     string       Optional. Resource ID of VPC which the ULB instances to list belong to 

  --subnet-id     string    Optional. Resource ID of subnet which the ULB instances to list
                            belong to 

  --group     string        Optional. Business group of ULB instances to list 

  --offset     int          Optional. Offset 

  --limit     int           Optional. Limit (default 50) 

  --help, -h                help for list 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud ulb <ucloud_ulb>` 	 - List and manipulate ULB instances

