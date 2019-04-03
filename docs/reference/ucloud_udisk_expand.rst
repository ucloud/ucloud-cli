.. _ucloud_udisk_expand:

ucloud udisk expand
-------------------

Expand udisk size

Synopsis
~~~~~~~~


Expand udisk size

::

  ucloud udisk expand [flags]

Options
~~~~~~~

::

  --udisk-id     strings    Required. Resource ID of the udisks to expand 

  --size-gb     int         Required. Size of the udisk after expanded. Unit: GB. Range [1,8000] 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone (default "cn-bj2-02") 

  --help, -h                help for expand 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud udisk <ucloud_udisk>` 	 - Read and manipulate udisk instances

