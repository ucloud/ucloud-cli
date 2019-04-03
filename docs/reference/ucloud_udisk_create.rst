.. _ucloud_udisk_create:

ucloud udisk create
-------------------

Create udisk instance

Synopsis
~~~~~~~~


Create udisk instance

::

  ucloud udisk create [flags]

Options
~~~~~~~

::

  --name     string              Required. Name of the udisk to create 

  --size-gb     int              Required. Size of the udisk to create. Unit:GB. Normal udisk
                                 [1,8000]; SSD udisk [1,4000]  (default 10) 

  --snapshot-id     string       Optional. Resource ID of a snapshot, which will apply to the
                                 udisk being created. If you set this option, 'udisk-type'
                                 will be omitted. 

  --project-id     string        Optional. Assign project-id (default "org-ryrmms") 

  --region     string            Optional. Assign region (default "cn-bj2") 

  --zone     string              Optional. Assign availability zone (default "cn-bj2-02") 

  --charge-type     string       Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay
                                 hourly (default "Dynamic") 

  --quantity     int             Optional. The duration of the instance. N years/months.
                                 (default 1) 

  --enable-data-ark     string   Optional. DataArk supports real-time backup, which can
                                 restore the udisk back to any moment within the last 12
                                 hours. (default "false") 

  --group     string             Optional. Business group (default "Default") 

  --udisk-type     string        Optional. 'Ordinary' or 'SSD' (default "Oridinary") 

  --async                        Optional. Do not wait for the long-running operation to finish. 

  --count     int                Optional. The count of udisk to create. Range [1,10] (default 1) 

  --help, -h                     help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud udisk <ucloud_udisk>` 	 - Read and manipulate udisk instances

