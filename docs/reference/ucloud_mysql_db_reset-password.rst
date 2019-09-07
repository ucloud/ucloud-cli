.. _ucloud_mysql_db_reset-password:

ucloud mysql db reset-password
------------------------------

Reset password of MySQL instances

Synopsis
~~~~~~~~


Reset password of MySQL instances

::

  ucloud mysql db reset-password [flags]

Options
~~~~~~~

::

  --udb-id     strings      Required. Resource ID of UDB instances to reset password 

  --password     string     Required. New password 

  --project-id     string   Optional. Override default project-id, see 'ucloud project list'
                            (default "org-ryrmms") 

  --region     string       Optional. Override default region, see 'ucloud region' (default
                            "cn-bj2") 

  --zone     string         Optional. Override default availability zone, see 'ucloud region'
                            (default "cn-bj2-02") 

  --help, -h                help for reset-password 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql db <ucloud_mysql_db>` 	 - Manange MySQL instances

