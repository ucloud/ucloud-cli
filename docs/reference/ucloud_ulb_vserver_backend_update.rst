.. _ucloud_ulb_vserver_backend_update:

ucloud ulb vserver backend update
---------------------------------

Update attributes of ULB backend nodes

Synopsis
~~~~~~~~


Update attributes of ULB backend nodes

::

  ucloud ulb vserver backend update [flags]

Options
~~~~~~~

::

  --ulb-id     string         Required. Resource ID of ULB which the backend nodes belong to 

  --backend-id     strings    Required. BackendID of backend nodes to update 

  --port     int              Optional. Port of your real server listening on backend nodes to
                              update. Rnage [1,65535] 

  --backend-mode     string   Optional. Enable backend node or not. Accept values: enable, disable 

  --weight     int            Optional. effective for lb-method WeightRoundrobin. Rnage
                              [0,100], -1 meaning no update (default -1) 

  --region     string         Optional. Override default region, see 'ucloud region' (default
                              "cn-bj2") 

  --project-id     string     Optional. Override default project-id, see 'ucloud project list'
                              (default "org-ryrmms") 

  --help, -h                  help for update 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud ulb vserver backend <ucloud_ulb_vserver_backend>` 	 - List and manipulate VServer backend nodes

