.. _ucloud_mysql_db_create:

ucloud mysql db create
----------------------

Create MySQL instance on UCloud platform

Synopsis
~~~~~~~~


Create MySQL instance on UCloud platform

::

  ucloud mysql db create [flags]

Options
~~~~~~~

::

  --project-id     string        Optional. Override default project-id, see 'ucloud project
                                 list' (default "org-ryrmms") 

  --region     string            Optional. Override default region, see 'ucloud region'
                                 (default "cn-bj2") 

  --zone     string              Optional. Override default availability zone, see 'ucloud
                                 region' (default "cn-bj2-02") 

  --version     string           Required. Version of udb instance 

  --name     string              Required. Name of udb instance to create, at least 6 letters 

  --conf-id     string           Required. ConfID of configuration. see 'ucloud mysql conf list' 

  --admin-user-name     string   Optional. Name of udb instance's administrator (default "root") 

  --password     string          Required. Password of udb instance's administrator 

  --backup-id     int            Optional. BackupID of the backup which the newly created UDB
                                 instance will recover from if specified. See 'ucloud mysql
                                 backup list' (default -1) 

  --port     int                 Optional. Port of udb instance (default 3306) 

  --disk-type     string         Optional. Setting this flag means using SSD disk. Accept
                                 values: 'normal','sata_ssd','pcie_ssd' 

  --disk-size-gb     int         Optional. Disk size of udb instance. From 20 to 3000
                                 according to memory size. Unit GB (default 20) 

  --memory-size-gb     int       Optional. Memory size of udb instance. From 1 to 128. Unit GB
                                 (default 1) 

  --mode     string              Optional. Mode of udb instance. Normal or HA, HA means
                                 high-availability. Both the normal and high-availability
                                 versions can create master-slave synchronization for data
                                 redundancy and read/write separation. The high-availability
                                 version provides a dual-master hot standby architecture to
                                 avoid database unavailability due to downtime or hardware
                                 failure. One more thing. It does better job for master-slave
                                 synchronization and disaster recovery using the InnoDB engine
                                 (default "Normal") 

  --vpc-id     string            Optional. Resource ID of VPC which the UDB to create belong
                                 to. See 'ucloud vpc list' 

  --subnet-id     string         Optional. Resource ID of subnet that the UDB to create belong
                                 to. See 'ucloud subnet list' 

  --async                        Optional. Do not wait for the long-running operation to finish. 

  --charge-type     string       Optional. Enumeration value.'Year',pay yearly;'Month',pay
                                 monthly; 'Dynamic', pay hourly; 'Trial', free trial(need
                                 permission) (default "Month") 

  --quantity     int             Optional. The duration of the instance. N years/months.
                                 (default 1) 

  --help, -h                     help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql db <ucloud_mysql_db>` 	 - Manange MySQL instances

