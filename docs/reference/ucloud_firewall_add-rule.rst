.. _ucloud_firewall_add-rule:

ucloud firewall add-rule
------------------------

Add rule to firewall instance

Synopsis
~~~~~~~~


Add rule to firewall instance

::

  ucloud firewall add-rule [flags]

Examples
~~~~~~~~

::

  ucloud firewall add-rule --fw-id firewall-2xxxxz/test.lxj2 --rules "TCP|24|0.0.0.0/0|ACCEPT|HIGH" --rules-file firewall_rules.txt

Options
~~~~~~~

::

  --fw-id     strings       Required. Resource ID of firewalls to update 

  --rules     strings       Required if rules-file is empay. Rules to add to firewall.
                            Schema:'Protocol|Port|IP|Action|Level'. See 'ucloud firewall
                            create --help' for detail. 

  --rules-file     string   Required if rules is empty. Path of rules file, in which each rule
                            occupies one line. Schema: Protocol|Port|IP|Action|Level. 

  --region     string       Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string   Optional. Project-id, see 'ucloud project list' (default "org-ryrmms") 

  --help, -h                help for add-rule 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud firewall <ucloud_firewall>` 	 - List and manipulate extranet firewall

