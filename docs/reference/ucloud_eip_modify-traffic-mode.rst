.. _ucloud_eip_modify-traffic-mode:

ucloud eip modify-traffic-mode
------------------------------

Modify charge mode of EIP instances

Synopsis
~~~~~~~~


Modify charge mode of EIP instances

::

  ucloud eip modify-traffic-mode [flags]

Examples
~~~~~~~~

::

  ucloud eip modify-traffic-mode --eip-id eip-xxx --traffic-mode Traffic

Options
~~~~~~~

::

  --eip-id     strings        Required, Resource ID of EIPs to modify charge mode 

  --traffic-mode     string   Required, Charge mode of eip, 'traffic' or 'bandwidth' 

  --project-id     string     Optional. Assign project-id (default "org-ryrmms") 

  --region     string         Optional. Assign region (default "cn-bj2") 

  --help, -h                  help for modify-traffic-mode 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud eip <ucloud_eip>` 	 - List,allocate and release EIP

