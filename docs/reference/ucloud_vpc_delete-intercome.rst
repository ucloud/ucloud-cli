.. _ucloud_vpc_delete-intercome:

ucloud vpc delete-intercome
---------------------------

delete the vpc intercome

Synopsis
~~~~~~~~


delete the vpc intercome

::

  ucloud vpc delete-intercome [flags]

Examples
~~~~~~~~

::

  ucloud vpc delete-intercome --vpc-id xxx --dst-vpc-id xxx

Options
~~~~~~~

::

  --vpc-id     string       Required. Resource ID of source VPC to disconnect with destination VPC 

  --dst-vpc-id     string   Required. Resource ID of destination VPC to disconnect with source VPC 

  --project-id     string   Optional. The project id of source vpc (default "org-ryrmms") 

  --region     string       Optional. The region of source vpc to disconnect (default "cn-bj2") 

  --dst-region     string   Optional. The region of dest vpc to disconnect 

  --help, -h                help for delete-intercome 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud vpc <ucloud_vpc>` 	 - List and manipulate VPC instances

