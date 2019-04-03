.. _ucloud_bw_shared_resize:

ucloud bw shared resize
-----------------------

Resize shared bandwidth instance's bandwidth

Synopsis
~~~~~~~~


Resize shared bandwidth instance's bandwidth

::

  ucloud bw shared resize [flags]

Options
~~~~~~~

::

  --shared-bw-id     string   Required. Resource ID of shared bandwidth instance to resize 

  --bandwidth-mb     int      Required. Unit:Mb. resize to bandwidth value 

  --region     string         Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string     Optional. Project-id, see 'ucloud project list' (default
                              "org-ryrmms") 

  --help, -h                  help for resize 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud bw shared <ucloud_bw_shared>` 	 - Create and manipulate shared bandwidth instances

