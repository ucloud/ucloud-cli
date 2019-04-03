.. _ucloud_udisk_attach:

ucloud udisk attach
-------------------

Attach udisk instances to an uhost

Synopsis
~~~~~~~~


Attach udisk instances to an uhost

::

  ucloud udisk attach [flags]

Examples
~~~~~~~~

::

  ucloud udisk attach --uhost-id uhost-xxxx --udisk-id bs-xxx1,bs-xxx2

Options
~~~~~~~

::

  --uhost-id     string     Required. Resource ID of the uhost instance which you want to
                            attach the disk 

  --udisk-id     strings    Required. Resource ID of the udisk instances to attach 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone (default "cn-bj2-02") 

  --async                   Optional. Do not wait for the long-running operation to finish. 

  --help, -h                help for attach 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud udisk <ucloud_udisk>` 	 - Read and manipulate udisk instances

