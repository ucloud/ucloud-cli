.. _ucloud_eip_bind:

ucloud eip bind
---------------

Bind EIP with uhost

Synopsis
~~~~~~~~


Bind EIP with uhost

::

  ucloud eip bind [flags]

Examples
~~~~~~~~

::

  ucloud eip bind --eip-id eip-xxx --resource-id uhost-xxx

Options
~~~~~~~

::

  --eip-id     string          Required. EIPId to bind 

  --resource-id     string     Required. ResourceID , which is the UHostId of uhost 

  --resource-type     string   Requried. ResourceType, type of resource to bind with eip.
                               'uhost','vrouter','ulb','upm','hadoophost'.eg.. (default "uhost") 

  --project-id     string      Optional. Assign project-id (default "org-ryrmms") 

  --region     string          Optional. Assign region (default "cn-bj2") 

  --help, -h                   help for bind 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud eip <ucloud_eip>` 	 - List,allocate and release EIP

