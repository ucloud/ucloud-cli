.. _ucloud_ulb_vserver_policy_list:

ucloud ulb vserver policy list
------------------------------

List content forward policies of the VServer instance

Synopsis
~~~~~~~~


List content forward policies of the VServer instance

::

  ucloud ulb vserver policy list [flags]

Options
~~~~~~~

::

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --ulb-id     string       Required. Resource ID of ULB 

  --vserver-id     string   Required. Resource ID of VServer 

  --help, -h                help for list 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud ulb vserver policy <ucloud_ulb_vserver_policy>` 	 - List and manipulate forward policy for VServer

