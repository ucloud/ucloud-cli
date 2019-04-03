.. _ucloud_uhost_resize:

ucloud uhost resize
-------------------

Resize uhost instance,such as cpu core count, memory size and disk size

Synopsis
~~~~~~~~


Resize uhost instance,such as cpu core count, memory size and disk size

::

  ucloud uhost resize [flags]

Examples
~~~~~~~~

::

  ucloud uhost resize --uhost-id uhost-xxx1,uhost-xxx2 --cpu 4 --memory-gb 8

Options
~~~~~~~

::

  --uhost-id     strings          Required. ResourceIDs(or UhostIDs) of the uhost instances 

  --project-id     string         Optional. Assign project-id (default "org-ryrmms") 

  --region     string             Optional. Assign region (default "cn-bj2") 

  --zone     string               Optional. Assign availability zone 

  --cpu     int                   Optional. The number of virtual CPU cores. Series1 {1, 2, 4,
                                  8, 12, 16, 24, 32}. Series2 {1,2,4,8,16} 

  --memory-gb     int             Optional. memory size. Unit: GB. Range: [1, 128], multiple of 2 

  --data-disk-size-gb     int     Optional. Data disk size,unit GB. Range[10,1000], SSD disk
                                  range[100,500]. Step 10 

  --system-disk-size-gb     int   Optional. System disk size, unit GB. Range[20,100]. Step 10.
                                  System disk does not support shrinkage 

  --net-cap     int               Optional. NIC scale. 1,upgrade; 2,downgrade; 0,unchanged 

  --yes, -y                       Optional. Do not prompt for confirmation. 

  --async, -a                     Optional. Do not wait for the long-running operation to finish. 

  --help, -h                      help for resize 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uhost <ucloud_uhost>` 	 - List,create,delete,stop,restart,poweroff or resize UHost instance

