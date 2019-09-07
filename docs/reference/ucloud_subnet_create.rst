.. _ucloud_subnet_create:

ucloud subnet create
--------------------

Create subnet of vpc network

Synopsis
~~~~~~~~


Create subnet of vpc network

::

  ucloud subnet create [flags]

Examples
~~~~~~~~

::

  ucloud subnet create --vpc-id uvnet-vpcxid --name testName --segment 192.168.2.0/24

Options
~~~~~~~

::

  --vpc-id     string       Required. Assign the VPC network of the subnet 

  --segment     ipNet       Required. Segment of subnet. For example '192.168.0.0/24' 

  --name     string         Optional. Name of subnet to create (default "Subnet") 

  --region     string       Optional. The region of the subnet (default "cn-bj2") 

  --project-id     string   Optional. The project id of the subnet (default "org-ryrmms") 

  --group     string        Optional. Business group 

  --remark     string       Optional. Remark of subnet to create 

  --help, -h                help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud subnet <ucloud_subnet>` 	 - List, create and delete subnet

