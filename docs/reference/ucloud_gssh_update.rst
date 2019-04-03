.. _ucloud_gssh_update:

ucloud gssh update
------------------

Update GlobalSSH instance

Synopsis
~~~~~~~~


Update GlobalSSH instance, including port and remark attribute

::

  ucloud gssh update [flags]

Examples
~~~~~~~~

::

  ucloud gssh update --gssh-id uga-xxx --port 22

Options
~~~~~~~

::

  --gssh-id     strings     Required. ResourceID of your GlobalSSH instances 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --port     int            Optional. Port of SSH service. 

  --remark     string       Optional. Remark of your GlobalSSH. 

  --help, -h                help for update 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud gssh <ucloud_gssh>` 	 - Create,list,update and delete globalssh instance

