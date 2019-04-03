.. _ucloud_redis_create:

ucloud redis create
-------------------

Create redis instance

Synopsis
~~~~~~~~


Create redis instance

::

  ucloud redis create [flags]

Options
~~~~~~~

::

  --name     string          Required. Name of the redis to create. Range of the password
                             length is [6,63] and the password can only contain letters and numbers 

  --type     string          Required. Type of the redis. Accept
                             values:'master-replica','distributed' 

  --size-gb     int          Optional. Memory size. Default value 1GB(for master-replica redis
                             type) or 16GB(for distributed redis type). Unit GB (default 1) 

  --version     string       Optional. Version of redis (default "3.2") 

  --vpc-id     string        Optional. VPC ID. This field is required under VPC2.0. See
                             'ucloud vpc list' 

  --subnet-id     string     Optional. Subnet ID. This field is required under VPC2.0. See
                             'ucloud subnet list' 

  --password     string      Optional. Password of redis to create 

  --region     string        Optional. Override default region, see 'ucloud region' (default
                             "cn-bj2") 

  --zone     string          Optional. Override default availability zone, see 'ucloud region'
                             (default "cn-bj2-02") 

  --project-id     string    Optional. Override default project-id, see 'ucloud project list'
                             (default "org-ryrmms") 

  --group     string         Optional. Business group 

  --charge-type     string   Optional. Enumeration value.'Year',pay yearly;'Month',pay
                             monthly; 'Dynamic', pay hourly; 'Trial', free trial(need
                             permission) (default "Month") 

  --quantity     int         Optional. The duration of the instance. N years/months. (default 1) 

  --help, -h                 help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud redis <ucloud_redis>` 	 - List and manipulate redis instances

