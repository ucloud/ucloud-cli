.. _ucloud_udisk_snapshot:

ucloud udisk snapshot
---------------------

Create shapshots for udisks

Synopsis
~~~~~~~~


Create shapshots for udisks

::

  ucloud udisk snapshot [flags]

Options
~~~~~~~

::

  --udisk-id     strings    Required. Resource ID of udisks to snapshot 

  --name     string         Required. Name of snapshots 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone (default "cn-bj2-02") 

  --comment     string      Optional. Description of snapshots 

  --async, -a               Optional. Do not wait for the long-running operation to finish. 

  --help, -h                help for snapshot 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud udisk <ucloud_udisk>` 	 - Read and manipulate udisk instances

