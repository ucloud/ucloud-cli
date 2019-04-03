.. _ucloud_memcache_create:

ucloud memcache create
----------------------

Create memcache instance

Synopsis
~~~~~~~~


Create memcache instance

::

  ucloud memcache create [flags]

Options
~~~~~~~

::

  --name     string          Required. Name of memcache instance to create 

  --size-gb     int          Optional. Memory size of memcache instance. Unit GB. Accpet
                             values:1,2,4,8,16,32 (default 1) 

  --vpc-id     string        Optional. VPC ID. See 'ucloud vpc list' 

  --subnet-id     string     Optional. Subnet ID. See 'ucloud subnet list' 

  --project-id     string    Optional. Override default project-id, see 'ucloud project list'
                             (default "org-ryrmms") 

  --region     string        Optional. Override default region, see 'ucloud region' (default
                             "cn-bj2") 

  --zone     string          Optional. Override default availability zone, see 'ucloud region'
                             (default "cn-bj2-02") 

  --charge-type     string   Optional. Enumeration value.'Year',pay yearly;'Month',pay
                             monthly; 'Dynamic', pay hourly; 'Trial', free trial(need
                             permission) (default "Month") 

  --quantity     int         Optional. The duration of the instance. N years/months. (default 1) 

  --group     string         Optional. Business group 

  --help, -h                 help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud memcache <ucloud_memcache>` 	 - List and manipulate memcache instances

