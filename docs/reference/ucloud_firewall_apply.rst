.. _ucloud_firewall_apply:

ucloud firewall apply
---------------------

Applay firewall to ucloud service

Synopsis
~~~~~~~~


Applay firewall to ucloud service

::

  ucloud firewall apply [flags]

Examples
~~~~~~~~

::

  ucloud firewall apply --fw-id firewall-xxx --resource-id uhost-xxx --resource-type uhost

Options
~~~~~~~

::

  --fw-id     string           Required. Resource ID of firewall to apply to some ucloud resource 

  --resource-type     string   Required. Resource type of resource to be applied firewall.
                               Range
                               'uhost','unatgw','upm','hadoophost','fortresshost','udhost','udockhost','dbaudit'. 

  --resource-id     strings    Resource ID of resources to be applied firewall 

  --region     string          Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string      Optional. Project-id, see 'ucloud project list' (default
                               "org-ryrmms") 

  --help, -h                   help for apply 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud firewall <ucloud_firewall>` 	 - List and manipulate extranet firewall

