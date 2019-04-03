.. _ucloud_uhost_stop:

ucloud uhost stop
-----------------

Shut down uhost instance

Synopsis
~~~~~~~~


Shut down uhost instance

::

  ucloud uhost stop [flags]

Examples
~~~~~~~~

::

  ucloud uhost stop --uhost-id uhost-xxx1,uhost-xxx2

Options
~~~~~~~

::

  --uhost-id     strings    Required. ResourceIDs(UHostIds) of the uhost instances 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone 

  --async                   Optional. Do not wait for the long-running operation to finish. 

  --help, -h                help for stop 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uhost <ucloud_uhost>` 	 - List,create,delete,stop,restart,poweroff or resize UHost instance

