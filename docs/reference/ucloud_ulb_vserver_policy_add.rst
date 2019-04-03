.. _ucloud_ulb_vserver_policy_add:

ucloud ulb vserver policy add
-----------------------------

Add content forward policy for VServer

Synopsis
~~~~~~~~


Add content forward policy for VServer

::

  ucloud ulb vserver policy add [flags]

Options
~~~~~~~

::

  --ulb-id     string           Required. Resource ID of ULB 

  --vserver-id     string       Required. Resource ID of VServer 

  --backend-id     strings      Required. BackendID of the VServer's backend nodes 

  --forward-method     string   Required. Forward method, accept values:Domain and Path; Both
                                forwarding methods can be described by using regular
                                expressions or wildcards 

  --expression     string       Required. Expression of domain or path, such as
                                "www.[123].demo.com" or "/path/img/*.jpg" 

  --region     string           Optional. Override default region, see 'ucloud region'
                                (default "cn-bj2") 

  --project-id     string       Optional. Override default project-id, see 'ucloud project
                                list' (default "org-ryrmms") 

  --help, -h                    help for add 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud ulb vserver policy <ucloud_ulb_vserver_policy>` 	 - List and manipulate forward policy for VServer

