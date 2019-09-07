.. _ucloud_uhost_reset-password:

ucloud uhost reset-password
---------------------------

Reset the administrator password for the UHost instances.

Synopsis
~~~~~~~~


Reset the administrator password for the UHost instances.

::

  ucloud uhost reset-password [flags]

Options
~~~~~~~

::

  --uhost-id     strings    Required. Resource IDs of the uhosts to reset the administrator's
                            password 

  --password     string     Required. New Password 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone (default "cn-bj2-02") 

  --yes, -y                 Optional. Do not prompt for confirmation. 

  --help, -h                help for reset-password 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uhost <ucloud_uhost>` 	 - List,create,delete,stop,restart,poweroff or resize UHost instance

