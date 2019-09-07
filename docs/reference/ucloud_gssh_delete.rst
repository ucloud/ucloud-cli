.. _ucloud_gssh_delete:

ucloud gssh delete
------------------

Delete GlobalSSH instance

Synopsis
~~~~~~~~


Delete GlobalSSH instance

::

  ucloud gssh delete [flags]

Examples
~~~~~~~~

::

  ucloud gssh delete --gssh-id uga-xx1  --id uga-xx2

Options
~~~~~~~

::

  --gssh-id     strings     Required. ID of the GlobalSSH instances you want to delete.
                            Multiple values specified by multiple commas 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --help, -h                help for delete 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud gssh <ucloud_gssh>` 	 - Create,list,update and delete globalssh instance

