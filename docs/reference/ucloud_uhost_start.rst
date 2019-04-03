.. _ucloud_uhost_start:

ucloud uhost start
------------------

Start Uhost instance

Synopsis
~~~~~~~~


Start Uhost instance

::

  ucloud uhost start [flags]

Examples
~~~~~~~~

::

  ucloud uhost start --uhost-id uhost-xxx1,uhost-xxx2

Options
~~~~~~~

::

  --uhost-id     strings    Requried. ResourceIDs(UHostIds) of the uhost instance 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone 

  --async                   Optional. Do not wait for the long-running operation to finish. 

  --help, -h                help for start 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uhost <ucloud_uhost>` 	 - List,create,delete,stop,restart,poweroff or resize UHost instance

