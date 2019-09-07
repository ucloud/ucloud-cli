.. _ucloud_uhost_reinstall-os:

ucloud uhost reinstall-os
-------------------------

Reinstall the operating system of the UHost instance

Synopsis
~~~~~~~~


Reinstall the operating system of the UHost instance. we will detach all udisk disks if the uhost attached some, and then stop the uhost if it's running

::

  ucloud uhost reinstall-os [flags]

Options
~~~~~~~

::

  --uhost-id     string     Required. Resource ID of the uhost to reinstall operating system 

  --password     string     Required. Password of the administrator 

  --image-id     string     Optional. Resource ID the image to install. See 'ucloud image
                            list'. Default is original image of the uhost 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone (default "cn-bj2-02") 

  --keep-data-disk          Keep data disk or not. If you keep data disk, you can't change OS
                            type(Linux->Window,e.g.) 

  --yes, -y                 Optional. Do not prompt for confirmation. 

  --async, -a               Optional. Do not wait for the long-running operation to finish. 

  --help, -h                help for reinstall-os 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uhost <ucloud_uhost>` 	 - List,create,delete,stop,restart,poweroff or resize UHost instance

