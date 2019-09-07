.. _ucloud_firewall_copy:

ucloud firewall copy
--------------------

Copy firewall

Synopsis
~~~~~~~~


Copy firewall

::

  ucloud firewall copy [flags]

Examples
~~~~~~~~

::

  ucloud firewall copy --src-fw firewall-xxx --target-region cn-bj2 --name test

Options
~~~~~~~

::

  --src-fw     string          Required. ResourceID or name of source firewall 

  --name     string            Required. Name of new firewall 

  --region     string          Optional. Current region, used to fetch source firewall
                               (default "cn-bj2") 

  --target-region     string   Optional. Copy firewall to target region (default "cn-bj2") 

  --project-id     string      Optional. Project-id, see 'ucloud project list' (default
                               "org-ryrmms") 

  --help, -h                   help for copy 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud firewall <ucloud_firewall>` 	 - List and manipulate extranet firewall

