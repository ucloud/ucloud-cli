.. _ucloud_mysql_db_create-slave:

ucloud mysql db create-slave
----------------------------

Create slave database

Synopsis
~~~~~~~~


Create slave database

::

  ucloud mysql db create-slave [flags]

Options
~~~~~~~

::

  --master-udb-id     string   Required. Resource ID of master UDB instance 

  --name     string            Required. Name of the slave DB to create 

  --port     int               Optional. Port of the slave db service (default 3306) 

  --region     string          Optional. Override default region, see 'ucloud region' (default
                               "cn-bj2") 

  --zone     string            Optional. Override default availability zone, see 'ucloud
                               region' (default "cn-bj2-02") 

  --project-id     string      Optional. Override default project-id, see 'ucloud project
                               list' (default "org-ryrmms") 

  --disk-type     string       Optional. Setting this flag means using SSD disk. Accept
                               values: normal, sata_ssd, pcie_ssd (default "Normal") 

  --memory-size-gb     int     Optional. Memory size of udb instance. From 1 to 128. Unit GB
                               (default 1) 

  --async                      Optional. Do not wait for the long-running operation to finish 

  --is-lock                    Optional. Lock master DB or not 

  --help, -h                   help for create-slave 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql db <ucloud_mysql_db>` 	 - Manange MySQL instances

