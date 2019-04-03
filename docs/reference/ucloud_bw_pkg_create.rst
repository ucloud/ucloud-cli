.. _ucloud_bw_pkg_create:

ucloud bw pkg create
--------------------

Create bandwidth package

Synopsis
~~~~~~~~


Create bandwidth package

::

  ucloud bw pkg create [flags]

Examples
~~~~~~~~

::

  ucloud bw pkg create --eip-id eip-xxx --bandwidth-mb 20 --start-time 2018-12-15/09:20:00 --end-time 2018-12-16/09:20:00

Options
~~~~~~~

::

  --eip-id     strings      Required. Resource ID of eip to be bound with created bandwidth package 

  --start-time     string   Required. The time to enable bandwidth package. Local time, for
                            example '2018-12-25/08:30:00' 

  --end-time     string     Required. The time to disable bandwidth package. Local time, for
                            example '2018-12-26/08:30:00' 

  --bandwidth-mb     int    Required. bandwidth of the bandwidth package to create.Range
                            [1,800]. Unit:'Mb'. 

  --region     string       Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string   Optional. Project-id, see 'ucloud project list' (default "org-ryrmms") 

  --help, -h                help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud bw pkg <ucloud_bw_pkg>` 	 - List, create and delete bandwidth package instances

