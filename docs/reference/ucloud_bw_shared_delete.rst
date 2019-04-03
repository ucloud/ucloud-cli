.. _ucloud_bw_shared_delete:

ucloud bw shared delete
-----------------------

Delete shared bandwidth instance

Synopsis
~~~~~~~~


Delete shared bandwidth instance

::

  ucloud bw shared delete [flags]

Options
~~~~~~~

::

  --shared-bw-id     strings   Required. Resource ID of shared bandwidth instances to delete 

  --eip-bandwidth-mb     int   Optional. Bandwidth of the joined EIPs,after deleting the
                               shared bandwidth instance (default 1) 

  --traffic-mode     string    Optional. The charge mode of joined EIPs after deleting the
                               shared bandwidth. Accept values:Bandwidth,Traffic 

  --region     string          Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string      Optional. Project-id, see 'ucloud project list' (default
                               "org-ryrmms") 

  --help, -h                   help for delete 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud bw shared <ucloud_bw_shared>` 	 - Create and manipulate shared bandwidth instances

