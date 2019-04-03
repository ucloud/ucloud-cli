.. _ucloud_image_copy:

ucloud image copy
-----------------

Copy custom images

Synopsis
~~~~~~~~


Copy custom images

::

  ucloud image copy [flags]

Options
~~~~~~~

::

  --source-image-id     strings    Required. Resource ID of source image 

  --project-id     string          Optional. Assign project-id (default "org-ryrmms") 

  --region     string              Optional. Assign region (default "cn-bj2") 

  --zone     string                Optional. Assign availability zone (default "cn-bj2-02") 

  --target-region     string       Optional. Target region. See 'ucloud region' (default "cn-bj2") 

  --target-project     string      Optional. Target Project ID. See 'ucloud project list'
                                   (default "org-ryrmms") 

  --target-image-name     string   Optional. Name of target image 

  --target-image-desc     string   Optional. Description of target image 

  --async                          Optional. Do not wait for the long-running operation to finish. 

  --help, -h                       help for copy 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud image <ucloud_image>` 	 - List and manipulate images

