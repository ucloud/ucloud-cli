.. _ucloud_mysql_db_resize:

ucloud mysql db resize
----------------------

Reszie MySQL instances, such as memory size, disk size and disk type

Synopsis
~~~~~~~~


Reszie MySQL instances, such as memory size, disk size and disk type

::

  ucloud mysql db resize [flags]

Options
~~~~~~~

::

  --udb-id     strings        Required. Resource ID of UDB instances to restart 

  --region     string         Optional. Override default region, see 'ucloud region' (default
                              "cn-bj2") 

  --zone     string           Optional. Override default availability zone, see 'ucloud
                              region' (default "cn-bj2-02") 

  --project-id     string     Optional. Override default project-id, see 'ucloud project list'
                              (default "org-ryrmms") 

  --memory-size-gb     int    Optional. Memory size of udb instance. From 1 to 128. Unit GB 

  --disk-size-gb     int      Optional. Disk size of udb instance. From 20 to 3000 according
                              to memory size. Unit GB. Step 10GB 

  --disk-type     string      Optional. Disk type of udb instance. Accept values:normal,
                              sata_ssd, pcie_ssd, normal_volume, sata_ssd_volume, pcie_ssd_volume 

  --start-after-upgrade       Optional. Automatic start the UDB instances after upgrade
                              (default true) 

  --async, -a                 Optional. Do not wait for the long-running operation to finish 

  --yes, -y                   Optional. Do not prompt for confirmation 

  --help, -h                  help for resize 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud mysql db <ucloud_mysql_db>` 	 - Manange MySQL instances

