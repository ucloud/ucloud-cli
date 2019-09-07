.. _ucloud_mysql_conf_clone:

ucloud mysql conf clone
-----------------------

Create configuration file by cloning existed configuration

Synopsis
~~~~~~~~


Create configuration file by cloning existed configuration

::

  ucloud mysql conf clone [flags]

Options
~~~~~~~

::

  --db-version     string    Required. Version of DB. Accept values:mysql-5.1, mysql-5.5,
                             mysql-5.6, mysql-5.7, percona-5.5, percona-5.6, percona-5.7,
                             mariadb-10.0 

  --name     string          Required. Name of configuration. It's length should be between 6
                             and 63 

  --description     string   Optional. Description of the configuration to clone (default " ") 

  --region     string        Optional. Override default region, see 'ucloud region' (default
                             "cn-bj2") 

  --zone     string          Optional. Override default availability zone, see 'ucloud region'
                             (default "cn-bj2-02") 

  --project-id     string    Optional. Override default project-id, see 'ucloud project list'
                             (default "org-ryrmms") 

  --src-conf-id     string   Optional. The ConfID of source configuration which to be cloned from 

  --help, -h                 help for clone 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql conf <ucloud_mysql_conf>` 	 - List and manipulate configuration files of MySQL instances

