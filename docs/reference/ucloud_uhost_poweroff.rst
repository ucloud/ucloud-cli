.. _ucloud_uhost_poweroff:

ucloud uhost poweroff
---------------------

Analog power off Uhost instnace

Synopsis
~~~~~~~~


Analog power off Uhost instnace

::

  ucloud uhost poweroff [flags]

Examples
~~~~~~~~

::

  ucloud uhost poweroff --uhost-id uhost-xxx1,uhost-xxx2

Options
~~~~~~~

::

  --uhost-id     strings    ResourceIDs(UHostIds) of the uhost instance 

  --project-id     string   Assign project-id (default "org-ryrmms") 

  --region     string       Assign region (default "cn-bj2") 

  --zone     string         Assign availability zone 

  --yes, -y                 Optional. Do not prompt for confirmation. 

  --help, -h                help for poweroff 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uhost <ucloud_uhost>` 	 - List,create,delete,stop,restart,poweroff or resize UHost instance

