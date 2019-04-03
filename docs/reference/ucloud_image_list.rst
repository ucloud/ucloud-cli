.. _ucloud_image_list:

ucloud image list
-----------------

List image

Synopsis
~~~~~~~~


List image

::

  ucloud image list [flags]

Examples
~~~~~~~~

::

  ucloud image list --image-type Base

Options
~~~~~~~

::

  --help, -h                help for list 

  --image-id     string     Optional. Resource ID of image 

  --image-type     string   Optional. 'Base',Standard image; 'Business',image market;
                            'Custom',custom image; Return all types by default 

  --limit     int           Optional. Max count (default 500) 

  --offset     int          Optional. Offset default 0 

  --os-type     string      Optional. Linux or Windows. Return all types by default 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud image <ucloud_image>` 	 - List and manipulate images

