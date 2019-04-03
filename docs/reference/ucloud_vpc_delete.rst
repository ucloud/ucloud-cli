.. _ucloud_vpc_delete:

ucloud vpc delete
-----------------

Delete vpc network

Synopsis
~~~~~~~~


Delete vpc network

::

  ucloud vpc delete [flags]

Examples
~~~~~~~~

::

  ucloud vpc delete --vpc-id uvnet-xxx

Options
~~~~~~~

::

  --vpc-id     strings      Required. Resource ID of the vpc network to delete 

  --region     string       Optional. Region of the vpc (default "cn-bj2") 

  --project-id     string   Optional. Project id of the vpc (default "org-ryrmms") 

  --help, -h                help for delete 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud vpc <ucloud_vpc>` 	 - List and manipulate VPC instances

