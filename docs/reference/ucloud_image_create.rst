.. _ucloud_image_create:

ucloud image create
-------------------

Create image from an uhost instance

Synopsis
~~~~~~~~


Create image from an uhost instance

::

  ucloud image create [flags]

Options
~~~~~~~

::

  --uhost-id     string     Resource ID of uhost to create image from 

  --image-name     string   Required. Name of the image to create 

  --image-desc     string   Optional. Description of the image to create 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone (default "cn-bj2-02") 

  --async, -a               Optional. Do not wait for the long-running operation to finish. 

  --help, -h                help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud image <ucloud_image>` 	 - List and manipulate images

