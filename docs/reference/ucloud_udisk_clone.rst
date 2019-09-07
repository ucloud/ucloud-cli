.. _ucloud_udisk_clone:

ucloud udisk clone
------------------

Clone an udisk

Synopsis
~~~~~~~~


Clone an udisk

::

  ucloud udisk clone [flags]

Options
~~~~~~~

::

  --source-id     string         Required. Resource ID of parent udisk 

  --name     string              Required. Name of new udisk 

  --project-id     string        Optional. Assign project-id (default "org-ryrmms") 

  --region     string            Optional. Assign region (default "cn-bj2") 

  --zone     string              Optional. Assign availability zone (default "cn-bj2-02") 

  --charge-type     string       Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay
                                 hourly (default "Month") 

  --quantity     int             Optional. The duration of the instance. N years/months.
                                 (default 1) 

  --enable-data-ark     string   Optional. DataArk supports real-time backup, which can
                                 restore the udisk back to any moment within the last 12
                                 hours. (default "false") 

  --coupon-id     string         Optional. Coupon ID, The Coupon can deduct part of the
                                 payment,see https://accountv2.ucloud.cn 

  --async                        Optional. Do not wait for the long-running operation to finish. 

  --help, -h                     help for clone 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud udisk <ucloud_udisk>` 	 - Read and manipulate udisk instances

