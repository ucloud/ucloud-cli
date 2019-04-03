.. _ucloud_bw_pkg_delete:

ucloud bw pkg delete
--------------------

Delete bandwidth packages

Synopsis
~~~~~~~~


Delete bandwidth packages

::

  ucloud bw pkg delete [flags]

Examples
~~~~~~~~

::

  ucloud bw pkg delete --resource-id bwpack-xxx

Options
~~~~~~~

::

  --resource-id     strings   Required, Resource ID of bandwidth package to delete 

  --region     string         Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string     Optional. Project-id, see 'ucloud project list' (default
                              "org-ryrmms") 

  --help, -h                  help for delete 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud bw pkg <ucloud_bw_pkg>` 	 - List, create and delete bandwidth package instances

