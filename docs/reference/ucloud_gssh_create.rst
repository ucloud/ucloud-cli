.. _ucloud_gssh_create:

ucloud gssh create
------------------

Create GlobalSSH instance

Synopsis
~~~~~~~~


Create GlobalSSH instance

::

  ucloud gssh create [flags]

Examples
~~~~~~~~

::

  ucloud gssh create --location Washington --target-ip 8.8.8.8

Options
~~~~~~~

::

  --location     string      Required. Location of the source server. See 'ucloud gssh location' 

  --target-ip     ip         Required. IP of the source server. Required 

  --project-id     string    Optional. Override default project-id, see 'ucloud project list'
                             (default "org-ryrmms") 

  --port     int             Optional. Port of The SSH service between 1 and 65535. Do not use
                             ports such as 80,443. (default 22) 

  --remark     string        Optional. Remark of your GlobalSSH. 

  --charge-type     string   Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay
                             hourly(requires access) (default "Month") 

  --quantity     int         Optional. The duration of the instance. N years/months. (default 1) 

  --help, -h                 help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud gssh <ucloud_gssh>` 	 - Create,list,update and delete globalssh instance

