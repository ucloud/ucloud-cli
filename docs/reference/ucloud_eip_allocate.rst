.. _ucloud_eip_allocate:

ucloud eip allocate
-------------------

Allocate EIP

Synopsis
~~~~~~~~


Allocate EIP

::

  ucloud eip allocate [flags]

Examples
~~~~~~~~

::

  ucloud eip allocate --line BGP --bandwidth-mb 2

Options
~~~~~~~

::

  --line     string                 Required. 'BGP' or 'International'. 'BGP' could be set in
                                    China mainland regions, such as cn-bj2 etc.
                                    'International' could be set in the regions beyond
                                    mainland, such as hk, tw-kh, us-ws etc. 

  --bandwidth-mb     int            Required. Bandwidth(Unit:Mbps).The range of value related
                                    to network charge mode. By traffic [1, 200]; by bandwidth
                                    [1,800] (Unit: Mbps); it could be 0 if the eip belong to
                                    the shared bandwidth 

  --project-id     string           Optional. Assign project-id (default "org-ryrmms") 

  --region     string               Optional. Assign region (default "cn-bj2") 

  --traffic-mode     string         Optional. traffic-mode is an enumeration value.
                                    'Traffic','Bandwidth' or 'ShareBandwidth' (default "Bandwidth") 

  --share-bandwidth-id     string   Optional. ShareBandwidthId, required only when
                                    traffic-mode is 'ShareBandwidth' 

  --quantity     int                Optional. The duration of the instance. N years/months.
                                    (default 1) 

  --charge-type     string          Optional. Enumeration value.'Year',pay yearly;'Month',pay
                                    monthly;'Dynamic', pay hourly(requires
                                    permission),'Trial', free trial(need permission) (default
                                    "Month") 

  --group     string                Optional. Group of your EIP. (default "Default") 

  --name     string                 Optional. Name of your EIP. (default "EIP") 

  --remark     string               Optional. Remark of your EIP. 

  --count     int                   Optional. Count of EIP to allocate (default 1) 

  --help, -h                        help for allocate 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud eip <ucloud_eip>` 	 - List,allocate and release EIP

