.. _ucloud_udpn_create:

ucloud udpn create
------------------

Create UDPN tunnel

Synopsis
~~~~~~~~


Create UDPN tunnel

::

  ucloud udpn create [flags]

Options
~~~~~~~

::

  --peer1     string         Required. One end of the tunnel to create (default "cn-bj2") 

  --peer2     string         Required. The other end of the tunnel create 

  --bandwidth-mb     int     Required. Bandwidth of the tunnel to create. Unit:Mb. Rnange [2,1000] 

  --charge-type     string   Optional. Enumeration value.'Year',pay yearly;'Month',pay
                             monthly;'Dynamic', pay hourly 

  --quantity     int         Optional. The duration of the instance. N years/months. (default 1) 

  --project-id     string    Optional. Project-id, see 'ucloud project list' (default "org-ryrmms") 

  --help, -h                 help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud udpn <ucloud_udpn>` 	 - List and manipulate udpn instances

