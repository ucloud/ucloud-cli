.. _ucloud_udpn_modify-bw:

ucloud udpn modify-bw
---------------------

Modify bandwidth of UDPN tunnel

Synopsis
~~~~~~~~


Modify bandwidth of UDPN tunnel

::

  ucloud udpn modify-bw [flags]

Options
~~~~~~~

::

  --udpn-id     strings     Required. Resource ID of UDPN to modify bandwidth 

  --bandwidth-mb     int    Required. Bandwidth of UDPN tunnel. Unit:Mb. Range [2,1000] 

  --region     string       Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string   Optional. Project-id, see 'ucloud project list' (default "org-ryrmms") 

  --help, -h                help for modify-bw 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud udpn <ucloud_udpn>` 	 - List and manipulate udpn instances

