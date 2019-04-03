.. _ucloud_mysql_db_promote-slave:

ucloud mysql db promote-slave
-----------------------------

Promote slave db to master

Synopsis
~~~~~~~~


Promote slave db to master

::

  ucloud mysql db promote-slave [flags]

Options
~~~~~~~

::

  --udb-id     strings      Required. Resource ID of slave db to promote 

  --is-force                Optional. Force to promote slave db or not. If the slave db falls
                            behind, the force promote may lose some data 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --help, -h                help for promote-slave 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql db <ucloud_mysql_db>` 	 - Manange MySQL instances

