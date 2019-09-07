.. _ucloud_vpc_create:

ucloud vpc create
-----------------

Create vpc network

Synopsis
~~~~~~~~


Create vpc network

::

  ucloud vpc create [flags]

Examples
~~~~~~~~

::

  ucloud vpc create --name xxx --segment 192.168.0.0/16

Options
~~~~~~~

::

  --name     string         Required. Name of the vpc network. 

  --segment     strings     Required. The segment for private network. 

  --group     string        Optional. Business group. 

  --remark     string       Optional. The description of the vpc. 

  --region     string       Optional. Assign the region of the VPC (default "cn-bj2") 

  --project-id     string   Optional. Assign the project-id (default "org-ryrmms") 

  --help, -h                help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud vpc <ucloud_vpc>` 	 - List and manipulate VPC instances

