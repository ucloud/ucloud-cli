.. _ucloud_udisk_list-snapshot:

ucloud udisk list-snapshot
--------------------------

List snaphosts

Synopsis
~~~~~~~~


List snaphosts

::

  ucloud udisk list-snapshot [flags]

Options
~~~~~~~

::

  --project-id     string     Optional. Assign project-id (default "org-ryrmms") 

  --region     string         Optional. Assign region (default "cn-bj2") 

  --zone     string           Optional. Assign availability zone (default "cn-bj2-02") 

  --snaphost-id     strings   Optional. Resource ID of snapshots to list 

  --uhost-id     string       Optional. Snapshots of the uhost 

  --disk-id     string        Optional. Snapshots of the udisk 

  --offset     int            Optional. Offset 

  --limit     int             Optional. Limit, length of snaphost list (default 50) 

  --help, -h                  help for list-snapshot 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud udisk <ucloud_udisk>` 	 - Read and manipulate udisk instances

