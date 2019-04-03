.. _ucloud_uhost_clone:

ucloud uhost clone
------------------

Create an uhost with the same configuration as another uhost, excluding bound eip and udisk

Synopsis
~~~~~~~~


Create an uhost with the same configuration as another uhost, excluding bound eip and udisk

::

  ucloud uhost clone [flags]

Options
~~~~~~~

::

  --uhost-id     string     Required. Resource ID of the uhost to clone from 

  --password     string     Required. Password of the uhost user(root/ubuntu) 

  --name     string         Optional. Name of the uhost to clone 

  --project-id     string   Optional. Assign project-id (default "org-ryrmms") 

  --region     string       Optional. Assign region (default "cn-bj2") 

  --zone     string         Optional. Assign availability zone (default "cn-bj2-02") 

  --async                   Optional. Do not wait for the long-running operation to finish. 

  --help, -h                help for clone 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uhost <ucloud_uhost>` 	 - List,create,delete,stop,restart,poweroff or resize UHost instance

