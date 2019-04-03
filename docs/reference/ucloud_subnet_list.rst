.. _ucloud_subnet_list:

ucloud subnet list
------------------

List subnet

Synopsis
~~~~~~~~


List subnet

::

  ucloud subnet list [flags]

Options
~~~~~~~

::

  --region     string       Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string   Optional. Project-id, see 'ucloud project list' (default "org-ryrmms") 

  --subnet-id     strings   Optional. Multiple values separated by commas 

  --vpc-id     string       Optional. Resource ID of VPC 

  --group     string        Optional. Group 

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

* :ref:`ucloud subnet <ucloud_subnet>` 	 - List, create and delete subnet

