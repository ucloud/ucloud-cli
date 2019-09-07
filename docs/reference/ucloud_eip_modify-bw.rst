.. _ucloud_eip_modify-bw:

ucloud eip modify-bw
--------------------

Modify bandwith of EIP instances

Synopsis
~~~~~~~~


Modify bandwith of EIP instances

::

  ucloud eip modify-bw [flags]

Examples
~~~~~~~~

::

  ucloud eip modify-bw --eip-id eip-xxx --bandwidth-mb 20

Options
~~~~~~~

::

  --eip-id     strings      Required, Resource ID of EIPs to modify bandwidth 

  --bandwidth-mb     int    Required. Bandwidth of EIP after modifed. Charge by traffic, range
                            [1,300]; charge by bandwidth, range [1,800] 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --help, -h                help for modify-bw 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud eip <ucloud_eip>` 	 - List,allocate and release EIP

