.. _ucloud_eip_leave-shared-bw:

ucloud eip leave-shared-bw
--------------------------

Leave shared bandwidth

Synopsis
~~~~~~~~


Leave shared bandwidth

::

  ucloud eip leave-shared-bw [flags]

Examples
~~~~~~~~

::

  ucloud eip leave-shared-bw --eip-id eip-b2gvu3

Options
~~~~~~~

::

  --eip-id     strings        Required. Resource ID of EIPs to leave shared bandwidth 

  --bandwidth-mb     int      Required. Bandwidth of EIP after leaving shared bandwidth,
                              ranging [1,300] for 'Traffic' charge mode, ranging [1,800] for
                              'Bandwidth' charge mode. Unit:Mb (default 1) 

  --traffic-mode     string   Optional. Charge mode of the EIP after leaving shared bandwidth,
                              'Bandwidth' or 'Traffic' (default "Bandwidth") 

  --shared-bw-id     string   Optional. Resource ID of shared bandwidth instance, assign this
                              flag to make the operation faster 

  --region     string         Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string     Optional. Project-id, see 'ucloud project list' (default
                              "org-ryrmms") 

  --help, -h                  help for leave-shared-bw 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud eip <ucloud_eip>` 	 - List,allocate and release EIP

