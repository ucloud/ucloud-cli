.. _ucloud_ulb_vserver_policy_update:

ucloud ulb vserver policy update
--------------------------------

Update content forward policies of ULB VServer

Synopsis
~~~~~~~~


Update content forward policies ULB VServer

::

  ucloud ulb vserver policy update [flags]

Options
~~~~~~~

::

  --region     string               Optional. Override default region, see 'ucloud region'
                                    (default "cn-bj2") 

  --project-id     string           Optional. Override default project-id, see 'ucloud project
                                    list' (default "org-ryrmms") 

  --ulb-id     string               Required. Resource ID of ULB 

  --vserver-id     string           Required. Resource ID of VServer 

  --policy-id     strings           Required. PolicyID of policies to update 

  --backend-id     strings          Optional. BackendID of backend nodes. If assign this flag,
                                    it will rewrite all backend nodes of the policy 

  --add-backend-id     strings      Optional. BackendID of backend nodes. Add backend nodes to
                                    the policy 

  --remove-backend-id     strings   Optional. BackendID of backend nodes. Remove those backend
                                    nodes from the policy 

  --forward-method     string       Optional. Forward method of policy, accept values:Domain
                                    and Path 

  --expression     string           Optional. Expression of domain or path, such as
                                    "www.[123].demo.com" or "/path/img/*.jpg" 

  --help, -h                        help for update 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud ulb vserver policy <ucloud_ulb_vserver_policy>` 	 - List and manipulate forward policy for VServer

