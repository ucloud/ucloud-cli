.. _ucloud_uhost_delete:

ucloud uhost delete
-------------------

Delete Uhost instance

Synopsis
~~~~~~~~


Delete Uhost instance

::

  ucloud uhost delete [flags]

Options
~~~~~~~

::

  --uhost-id     strings    Requried. ResourceIDs(UhostIds) of the uhost instance 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --zone     string         Optional. availability zone 

  --destory                 Optional. false,the uhost instance will be thrown to UHost recycle
                            if you have permission; true,the uhost instance will be deleted
                            directly 

  --release-eip             Optional. false,Unbind EIP only; true, Unbind EIP and release it 

  --delete-cloud-disk       Optional. false,Detach cloud disk only; true, Detach cloud disk
                            and delete it 

  --yes, -y                 Optional. Do not prompt for confirmation. 

  --help, -h                help for delete 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uhost <ucloud_uhost>` 	 - List,create,delete,stop,restart,poweroff or resize UHost instance

