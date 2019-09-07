.. _ucloud_bw_shared_create:

ucloud bw shared create
-----------------------

Create shared bandwidth instance

Synopsis
~~~~~~~~


Create shared bandwidth instance

::

  ucloud bw shared create [flags]

Options
~~~~~~~

::

  --name     string          Required. Name of the shared bandwidth instance 

  --bandwidth-mb     int     Optional. Unit:Mb. Bandwidth of the shared bandwidth. Range
                             [20,5000] (default 20) 

  --region     string        Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string    Optional. Project-id, see 'ucloud project list' (default "org-ryrmms") 

  --charge-type     string   Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay
                             hourly (default "Month") 

  --quantity     int         Optional. The duration of the instance. N years/months. (default 1) 

  --help, -h                 help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud bw shared <ucloud_bw_shared>` 	 - Create and manipulate shared bandwidth instances

